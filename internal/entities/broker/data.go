package broker

import "github.com/google/uuid"

type BrokerData struct {
	CommName    string
	ChatID      int64
	Value       []byte
	MessageUuid uuid.UUID
	IsFile      bool
}
