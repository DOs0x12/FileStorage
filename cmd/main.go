package main

import (
	"context"
	"flag"

	"github.com/DOs0x12/TeleBot/client/v2/broker"
	"github.com/sirupsen/logrus"

	"github.com/DOs0x12/FileStorage/internal/infrastructure/config"
)

func main() {
	configPath := flag.String("conf", "../etc/config.yml", "Config path.")
	flag.Parse()

	conf, err := config.LoadConfig(*configPath)
	if err != nil {
		logrus.Error("Failed to load the config: ", err)

		return
	}

	appCtx := context.Background()
	serviceName := "storage-of-receipts"
	broker, err := broker.NewKafkaBroker(appCtx, conf.KafkaAddress, serviceName)
	if err != nil {
		logrus.Error("Failed to create a broker: ", err)

		return
	}

}
