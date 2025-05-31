package app

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	brokerEnt "github.com/DOs0x12/FileStorage/internal/entities/broker"
	brokerInt "github.com/DOs0x12/FileStorage/internal/interfaces/broker"
	fileInt "github.com/DOs0x12/FileStorage/internal/interfaces/file"
	storageInt "github.com/DOs0x12/FileStorage/internal/interfaces/storage"
	"github.com/DOs0x12/TeleBot/server/v3/tmp_storage"
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
	sessionSt := tmp_storage.NewTmpStorage()
	sLifetime := 2 * time.Hour
	sClPer := 1 * time.Hour
	sessionSt.StartCleanupOldObjs(ctx, sLifetime, sClPer)
	for d := range dataChan {
		if d.CommName == SendComm {
			processSendingState(ctx, d, servSet, sessionSt)
			commitMsg(ctx, d.MessageUuid, servSet.Broker)

			continue
		}

		if d.CommName == GetComm {
			processGettingState(ctx, d, servSet, sessionSt)
			commitMsg(ctx, d.MessageUuid, servSet.Broker)

			continue
		}

		if d.CommName == GetAllComm {
			processGettingFiles(ctx, d, servSet)
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

const failedProcErr = "Не удалось обработать файл"

func processSendingState(
	ctx context.Context,
	brokerData brokerEnt.BrokerData,
	servSet ServiceSet,
	sessionSt tmp_storage.TmpStorage,
) {
	chatUuid := int64ToUUID(brokerData.ChatID)
	stObj, ok := sessionSt.GetObj(chatUuid)
	var currSt state
	rawCommName := []byte("/" + brokerData.CommName)
	if !ok || slices.Equal(brokerData.Value, rawCommName) {
		currSt = comm
	} else {
		currSt = stObj.Obj.(state)
	}

	const sendingFileMessage = "Отправь файл(ы) для загрузки в хранилище"
	const noFileUserErr = "Сообщение не содержит файл"

	switch currSt {
	case comm:
		if sendMsgWithErrHandling(ctx, brokerData, servSet.Broker, sendingFileMessage) {
			sessionSt.AddObjByUuid(data, chatUuid)
		}
	case data:
		if !brokerData.IsFile {
			logrus.Error("An incoming message has no file")
			_ = sendMsgWithErrHandling(ctx, brokerData, servSet.Broker, noFileUserErr)

			return
		}

		err := processFileData(ctx, brokerData.Value, servSet)
		if err != nil {
			logrus.Error("Failed to process file data: ", err)
			_ = sendMsgWithErrHandling(ctx, brokerData, servSet.Broker, failedProcErr)
		}
	}
}

func int64ToUUID(id int64) uuid.UUID {
	idBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(idBytes, uint64(id))
	namespace := uuid.NameSpaceOID // Can be any UUID
	return uuid.NewSHA1(namespace, idBytes)
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
		Value:    []byte(msg),
	}
	err := broker.SendData(ctx, d)
	if err != nil {
		logrus.Error("Failed to send a message to the broker: ", err)

		return false
	}

	return true
}

type FileDto struct {
	Name string
	Data []byte
}

func processFileData(
	ctx context.Context,
	rawData []byte,
	servSet ServiceSet,
) error {
	if len(rawData) == 0 {
		return errors.New("data is empty")
	}

	var dto FileDto
	err := json.Unmarshal(rawData, &dto)
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

	err = servSet.Storage.Insert(ctx, fNum, fName)
	if err != nil {
		return err
	}

	err = servSet.File.Write(dto.Data, fName)
	if err != nil {
		refErr := servSet.Storage.DeleteReference(ctx, fNum)
		if refErr != nil {
			return fmt.Errorf(
				"cannot delete the reference with the number %v after failing to write file data: %w", fNum, err)
		}

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
) error {
	fd, name, err := getFileData(ctx, st, file, brData.Value)
	if err != nil {
		return err
	}

	dto := FileDto{Name: name, Data: fd}
	d, err := json.Marshal(dto)
	if err != nil {
		return fmt.Errorf("failed to marshal a file DTO for the broker: %w", err)
	}

	dataToSend := brokerEnt.BrokerData{CommName: brData.CommName, ChatID: brData.ChatID, Value: d, IsFile: true}

	err = br.SendData(ctx, dataToSend)
	if err != nil {
		return fmt.Errorf("failed to send data to the broker: %w", err)
	}

	return nil
}

func getFileData(
	ctx context.Context,
	st storageInt.ReferenceStorage,
	file fileInt.File,
	rawData []byte,
) ([]byte, string, error) {
	id, err := strconv.ParseInt(string(rawData), 0, 64)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get the ID of a file reference from a broker data: %w", err)
	}

	ref, err := st.GetReference(ctx, id)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get the file: %w", err)
	}

	fd, err := file.Read(ref)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get the file data: %w", err)
	}

	return fd, ref, nil
}

func processGettingState(
	ctx context.Context,
	brokerData brokerEnt.BrokerData,
	servSet ServiceSet,
	sessionSt tmp_storage.TmpStorage,
) {
	chatUuid := int64ToUUID(brokerData.ChatID)
	stObj, ok := sessionSt.GetObj(chatUuid)
	var currSt state
	if !ok {
		currSt = comm
	} else {
		currSt = stObj.Obj.(state)
	}

	const sendingFileMessage = "Отправь номер файла"

	switch currSt {
	case comm:
		if sendMsgWithErrHandling(ctx, brokerData, servSet.Broker, sendingFileMessage) {
			sessionSt.AddObjByUuid(data, chatUuid)
		}
	case data:
		if err := processFile(ctx, servSet.Storage, servSet.File, servSet.Broker, brokerData); err != nil {
			logrus.Error(err)
			_ = sendMsgWithErrHandling(ctx, brokerData, servSet.Broker, failedProcErr)

			return
		}
		sessionSt.DelObj(chatUuid)
	}
}

func processGettingFiles(
	ctx context.Context,
	brokerData brokerEnt.BrokerData,
	servSet ServiceSet,
) {
	const failedGetAllRefErr = "Не удалось получить данные всех файлов"
	refs, err := servSet.Storage.GetAllReferences(ctx)
	if err != nil {
		logrus.Error("Failed to get all references: ", err)
		_ = sendMsgWithErrHandling(ctx, brokerData, servSet.Broker, failedGetAllRefErr)

		return
	}

	msg := strings.Join(refs, "\n")
	_ = sendMsgWithErrHandling(ctx, brokerData, servSet.Broker, msg)
}
