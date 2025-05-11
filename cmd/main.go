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

	cn := "send"
	cd := "Send a receipt file"
	comm := brokerEnt.CommandData{Name: cn, Description: cd}
	err = broker.RegisterCommand(appCtx, comm, serviceName)
	if err != nil {
		logrus.Errorf("Failed to register a command %v in the bot: %v", cn, err)

		return
	}

	wr := fileInfra.NewWriter(*folderPath)
	ext := fileDom.Extractor{}

	app.Serve(appCtx, broker, wr, ext)
}
