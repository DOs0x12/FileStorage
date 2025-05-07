package main

import (
	"context"
	"flag"

	"github.com/sirupsen/logrus"

	infraBroker "github.com/DOs0x12/FileStorage/internal/infrastructure/broker"
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
	broker, err := infraBroker.NewBroker(appCtx, conf.KafkaAddress, serviceName)
	if err != nil {
		logrus.Error("Failed to create a broker: ", err)

		return
	}

}
