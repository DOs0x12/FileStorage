package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	brokerEnt "github.com/DOs0x12/FileStorage/internal/entities/broker"
	brokerInt "github.com/DOs0x12/FileStorage/internal/interfaces/broker"
	fileInt "github.com/DOs0x12/FileStorage/internal/interfaces/file"
	storageInt "github.com/DOs0x12/FileStorage/internal/interfaces/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type ServiceSet struct {
	Broker    brokerInt.MessageBroker
	File      fileInt.File
	Extractor fileInt.Extractor
	Storage   storageInt.ReferenceStorage
}

var (
	SendComm   = "sendf"
	GetComm    = "getf"
	GetAllComm = "getallf"
)

func Serve(ctx context.Context, servSet ServiceSet) {
	dataChan := servSet.Broker.StartGetData(ctx)

	for d := range dataChan {
		if d.CommName == SendComm {
			processSendingState(ctx, d, servSet)
			commitMsg(ctx, d.MessageUuid, servSet.Broker)

			continue
		}

		if d.CommName == GetComm {
			processGettingState(ctx, d, servSet)
			commitMsg(ctx, d.MessageUuid, servSet.Broker)

			continue
		}

		if d.CommName == GetComm {
			processGettingState(ctx, d, servSet)
			commitMsg(ctx, d.MessageUuid, servSet.Broker)
		}
	}

	servSet.Broker.Stop()
	logrus.Info("The application was stopped")
}

func commitMsg(ctx context.Context, uuid uuid.UUID, br brokerInt.MessageBroker) {
	err := br.Commit(ctx, uuid)
	if err != nil {
		logrus.Errorf("Failed to commit the messsage with UUID: %v: %v", uuid.String(), err)
	}
}

type state int

const (
	comm state = iota
	data
)

var sendingSessions = make(map[int64]state)

func processSendingState(
	ctx context.Context,
	brokerData brokerEnt.BrokerData,
	servSet ServiceSet,
) {
	currSt, ok := sendingSessions[brokerData.ChatID]
	if !ok {
		currSt = comm
	}

	const sendingFileMessage = "Отправь файл для загрузки в хранилище"

	switch currSt {
	case comm:
		if sendMsgWithErrHandling(ctx, brokerData, servSet.Broker, sendingFileMessage) {
			sendingSessions[brokerData.ChatID] = data
		}
	case data:
		if !brokerData.IsFile {
			logrus.Error("An incoming message has no file")

			return
		}

		err := processFileData(ctx, brokerData.Value, servSet)
		if err != nil {
			logrus.Error("Failed to process file data: ", err)
		}
		delete(sendingSessions, brokerData.ChatID)
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

type FileDto struct {
	Name,
	Data string
}

func processFileData(
	ctx context.Context,
	rawData string,
	servSet ServiceSet,
) error {
	if rawData == "" {
		return errors.New("data is empty")
	}

	var dto FileDto
	err := json.Unmarshal([]byte(rawData), &dto)
	if err != nil {
		return fmt.Errorf("failed to unmarshal file data: %w", err)
	}

	fName := servSet.Extractor.ExtractFileName(dto.Name)
	if fName == "" {
		return fmt.Errorf("file name '%v' in wrong format", dto.Name)
	}

	fNum, err := servSet.Extractor.ExtractNumber(dto.Name)
	if err != nil {
		return err
	}

	err = servSet.File.Write(dto.Data, fName)
	if err != nil {
		return err
	}

	err = servSet.Storage.Insert(ctx, fNum, fName)
	if err != nil {
		return err
	}

	return nil
}

func processFile(
	ctx context.Context,
	st storageInt.ReferenceStorage,
	file fileInt.File,
	br brokerInt.MessageBroker,
	brData brokerEnt.BrokerData,
) {
	fd, name, err := getFileData(ctx, st, file, brData.Value)
	if err != nil {
		logrus.Error(err)

		return
	}

	dto := FileDto{Name: name, Data: fd}
	d, err := json.Marshal(dto)
	if err != nil {
		logrus.Error("failed to marshal a file DTO for the broker: ", err)

		return
	}

	dataToSend := brokerEnt.BrokerData{CommName: brData.CommName, ChatID: brData.ChatID, Value: string(d), IsFile: true}

	err = br.SendData(ctx, dataToSend)
	if err != nil {
		logrus.Error("failed to send data to the broker: ", err)
	}
}

func getFileData(
	ctx context.Context,
	st storageInt.ReferenceStorage,
	file fileInt.File,
	rawData string,
) (string, string, error) {
	id, err := strconv.ParseInt(rawData, 0, 64)
	if err != nil {
		return "", "", fmt.Errorf("failed to get the ID of a file reference from a broker data: %w", err)
	}

	ref, err := st.GetReference(ctx, id)
	if err != nil {
		return "", "", fmt.Errorf("failed to get the file: %w", err)
	}

	fd, err := file.Read(ref)
	if err != nil {
		return "", "", fmt.Errorf("failed to get the file data: %w", err)
	}

	return fd, ref, nil
}

var gettingSessions = make(map[int64]state)

func processGettingState(
	ctx context.Context,
	brokerData brokerEnt.BrokerData,
	servSet ServiceSet,
) {
	currSt, ok := gettingSessions[brokerData.ChatID]
	if !ok {
		currSt = comm
	}

	const sendingFileMessage = "Отправь номер файла"

	switch currSt {
	case comm:
		if sendMsgWithErrHandling(ctx, brokerData, servSet.Broker, sendingFileMessage) {
			gettingSessions[brokerData.ChatID] = data
		}
	case data:
		processFile(ctx, servSet.Storage, servSet.File, servSet.Broker, brokerData)
		delete(gettingSessions, brokerData.ChatID)
	}
}
