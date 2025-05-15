package main

import (
	"context"
	"flag"

	"github.com/sirupsen/logrus"

	"github.com/DOs0x12/FileStorage/internal/app"
	fileDom "github.com/DOs0x12/FileStorage/internal/domain/file"
	brokerEnt "github.com/DOs0x12/FileStorage/internal/entities/broker"
	brokerInfra "github.com/DOs0x12/FileStorage/internal/infrastructure/broker"
	"github.com/DOs0x12/FileStorage/internal/infrastructure/config"
	fileInfra "github.com/DOs0x12/FileStorage/internal/infrastructure/file"
	storageInfra "github.com/DOs0x12/FileStorage/internal/infrastructure/storage"
)

func main() {
	configPath := flag.String("conf", "../etc/config.yml", "Config path.")
	folderPath := flag.String("folder", "", "File storage folder path.")
	flag.Parse()

	if *folderPath == "" {
		flag.PrintDefaults()

		return
	}

	conf, err := config.LoadConfig(*configPath)
	if err != nil {
		logrus.Error("Failed to load the config: ", err)

		return
	}

	appCtx := context.Background()
	serviceName := "receipt-storage"
	broker, err := brokerInfra.NewBroker(appCtx, conf.KafkaAddress, serviceName)
	if err != nil {
		logrus.Error("Failed to create a broker: ", err)

		return
	}

	sendComnName := app.SendComm
	sendCommDesc := "Send a receipt file"

	comm := brokerEnt.CommandData{Name: sendComnName, Description: sendCommDesc}
	err = broker.RegisterCommand(appCtx, comm, serviceName)
	if err != nil {
		logrus.Errorf("Failed to register a command %v in the bot: %v", sendComnName, err)

		return
	}

	getCommName := app.GetComm
	getCommDesc := "Get a receipt file"

	comm = brokerEnt.CommandData{Name: getCommName, Description: getCommDesc}
	err = broker.RegisterCommand(appCtx, comm, serviceName)
	if err != nil {
		logrus.Errorf("Failed to register a command %v in the bot: %v", getCommName, err)

		return
	}

	getAllCommName := app.GetAllComm
	getAllCommDesc := "Get all files with numbers"

	comm = brokerEnt.CommandData{Name: getAllCommName, Description: getAllCommDesc}
	err = broker.RegisterCommand(appCtx, comm, serviceName)
	if err != nil {
		logrus.Errorf("Failed to register a command %v in the bot: %v", getAllCommName, err)

		return
	}

	wr := fileInfra.NewWriter(*folderPath)
	ext := fileDom.Extractor{}

	stConf := storageInfra.StorageConf{
		Address:  conf.StorageAddress,
		Database: conf.StorageDB,
		User:     conf.StorageUser,
		Pass:     conf.StoragePass,
	}
	st, err := storageInfra.NewPgRefStorage(appCtx, stConf)
	if err != nil {
		logrus.Error("Failed to connect to a storage: ", err)

		return
	}

	sSet := app.ServiceSet{Broker: broker, File: wr, Extractor: ext, Storage: st}
	app.Serve(appCtx, sSet)
}
