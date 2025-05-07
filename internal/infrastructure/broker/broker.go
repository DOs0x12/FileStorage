package broker

import (
	"context"
	"fmt"

	brokerEnt "github.com/DOs0x12/FileStorage/internal/entities/broker"
	"github.com/DOs0x12/TeleBot/client/v2/broker"
	"github.com/google/uuid"
)

type MessageBroker struct {
	kafkaBroker *broker.KafkaBroker
}

func NewBroker(ctx context.Context, address, serviceName string) (MessageBroker, error) {
	br, err := broker.NewKafkaBroker(ctx, address, serviceName)
	if err != nil {
		return MessageBroker{}, fmt.Errorf("failed to create a kafka broker: %w", err)
	}

	return MessageBroker{kafkaBroker: br}, nil
}

func (b MessageBroker) StartGetData(ctx context.Context) <-chan brokerEnt.BrokerData {
	kBRDataChan := b.kafkaBroker.StartGetData(ctx)
	brDataChan := make(chan brokerEnt.BrokerData)

	go piplineBrockerData(kBRDataChan, brDataChan)

	return brDataChan
}

func piplineBrockerData(
	kBRDataChan <-chan broker.BrokerData,
	brDataChan chan<- brokerEnt.BrokerData,
) {
	defer close(brDataChan)
	for d := range kBRDataChan {
		brDataChan <- brokerEnt.BrokerData{
			CommName:    d.CommName,
			ChatID:      d.ChatID,
			Value:       d.Value,
			MessageUuid: d.MessageUuid,
		}
	}
}

func (b MessageBroker) Commit(ctx context.Context, msgUuid uuid.UUID) error {
	return b.kafkaBroker.Commit(ctx, msgUuid)
}

func (b MessageBroker) SendData(ctx context.Context, data brokerEnt.BrokerData) error {
	kBRData := broker.BrokerData{
		CommName:    data.CommName,
		ChatID:      data.ChatID,
		Value:       data.Value,
		MessageUuid: data.MessageUuid,
	}

	return b.kafkaBroker.SendData(ctx, kBRData)
}

func (b MessageBroker) RegisterCommand(
	ctx context.Context,
	commData brokerEnt.CommandData,
	serviceName string,
) error {
	kBRComData := broker.BrokerCommandData{Name: commData.Name, Description: commData.Description}

	return b.kafkaBroker.RegisterCommand(ctx, kBRComData, serviceName)
}

func (b MessageBroker) Stop() {
	b.kafkaBroker.Stop()
}
