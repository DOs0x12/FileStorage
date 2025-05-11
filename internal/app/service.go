package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	brokerEnt "github.com/DOs0x12/FileStorage/internal/entities/broker"
	brokerInt "github.com/DOs0x12/FileStorage/internal/interfaces/broker"
	fileInt "github.com/DOs0x12/FileStorage/internal/interfaces/file"
	storageInt "github.com/DOs0x12/FileStorage/internal/interfaces/storage"
	"github.com/sirupsen/logrus"
)

func Serve(
	ctx context.Context,
	broker brokerInt.MessageBroker,
	fileWriter fileInt.Writer,
	extractor fileInt.Extractor,
	storage storageInt.ReferenceStorage,
) {
	dataChan := broker.StartGetData(ctx)

	for d := range dataChan {
		processState(ctx, d, broker, fileWriter, extractor, storage)
		err := broker.Commit(ctx, d.MessageUuid)
		if err != nil {
			logrus.Errorf("Failed to commit the messsage with UUID: %v: %v", d.MessageUuid.String(), err)
		}
	}

	broker.Stop()
	logrus.Info("The application was stopped")
}

type state int

const (
	comm state = iota
	data
)

var sessions = make(map[int64]state)

type FileDto struct {
	Name,
	Data string
}

func processState(
	ctx context.Context,
	brokerData brokerEnt.BrokerData,
	broker brokerInt.MessageBroker,
	fileWriter fileInt.Writer,
	extractor fileInt.Extractor,
	storage storageInt.ReferenceStorage,
) {
	currSt, ok := sessions[brokerData.ChatID]
	if !ok {
		currSt = comm
	}

	const sendingFileMessage = "Отправь файл для загрузки в хранилище"

	switch currSt {
	case comm:
		if sendMsgWithErrHandling(ctx, brokerData, broker, sendingFileMessage) {
			sessions[brokerData.ChatID] = data
		}
	case data:
		err := processFileData(ctx, brokerData.Value, fileWriter, extractor, storage)
		if err != nil {
			logrus.Error("Failed to process file data: ", err)
		}
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

func processFileData(
	ctx context.Context,
	rawData string,
	fileWriter fileInt.Writer,
	extractor fileInt.Extractor,
	storage storageInt.ReferenceStorage,
) error {
	if rawData == "" {
		return errors.New("data is empty")
	}

	var dto FileDto
	err := json.Unmarshal([]byte(rawData), &dto)
	if err != nil {
		return fmt.Errorf("failed to unmarshal file data: %w", err)
	}

	fName := extractor.ExtractFileName(dto.Name)
	if fName == "" {
		return fmt.Errorf("file name '%v' in wrong format", dto.Name)
	}

	fNum, err := extractor.ExtractNumber(dto.Name)
	if err != nil {
		return err
	}

	err = fileWriter.Write(dto.Data, fName)
	if err != nil {
		return err
	}

	err = storage.Insert(ctx, fNum, fName)
	if err != nil {
		return err
	}

	return nil
}
