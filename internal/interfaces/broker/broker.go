package broker

import (
	"context"

	brokerEnt "github.com/DOs0x12/FileStorage/internal/entities/broker"
	"github.com/google/uuid"
)

type MessageBroker interface {
	StartGetData(ctx context.Context) <-chan brokerEnt.BrokerData
	Commit(ctx context.Context, msgUuid uuid.UUID) error
	SendData(ctx context.Context, data brokerEnt.BrokerData) error
	Stop()
}
