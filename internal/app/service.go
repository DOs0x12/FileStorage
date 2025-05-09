package app

import (
	"context"

	brokerEnt "github.com/DOs0x12/FileStorage/internal/entities/broker"
	brokerInt "github.com/DOs0x12/FileStorage/internal/interfaces/broker"
	"github.com/sirupsen/logrus"
)

func Serve(ctx context.Context, broker brokerInt.MessageBroker) {
	dataChan := broker.StartGetData(ctx)

	for {
		select {
		case d := <-dataChan:
			processState(ctx, d, broker)
			err := broker.Commit(ctx, d.MessageUuid)
			if err != nil {
				logrus.Errorf("Failed to commit the messsage with UUID: %v: %v", d.MessageUuid.String(), err)
			}
		case <-ctx.Done():
			broker.Stop()
			logrus.Info("The application was stopped")
		}
	}
}

type state int

const (
	comm state = iota
	data
	name
)

var sessions = make(map[int64]state)

func processState(ctx context.Context, brokerData brokerEnt.BrokerData, broker brokerInt.MessageBroker) {
	currSt, ok := sessions[brokerData.ChatID]
	if !ok {
		currSt = comm
	}

	const (
		sendingFileMessage = "Отправь файл для загрузки в хранилище"
		fileNameMessage    = "Напиши имя для объекта в хранилище"
	)

	switch currSt {
	case comm:
		if sendMsgWithErrHandling(ctx, brokerData, broker, sendingFileMessage) {
			sessions[brokerData.ChatID] = data
		}
	case data:
		if sendMsgWithErrHandling(ctx, brokerData, broker, fileNameMessage) {
			sessions[brokerData.ChatID] = name
		}
	case name:
		delete(sessions, brokerData.ChatID)
	}
}

func sendMsgWithErrHandling(
	ctx context.Context,
	brokerData brokerEnt.BrokerData,
	broker brokerInt.MessageBroker,
	msg string,
) bool {
	d := brokerEnt.BrokerData{
		CommName: brokerData.CommName,
		ChatID:   brokerData.ChatID,
		Value:    msg,
	}
	err := broker.SendData(ctx, d)
	if err != nil {
		logrus.Error("Failed to send a message to the broker: ", err)

		return false
	}

	return true
}
