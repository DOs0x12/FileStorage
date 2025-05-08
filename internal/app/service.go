package app

import (
	"context"

	"github.com/DOs0x12/FileStorage/internal/interfaces/broker"
	"github.com/sirupsen/logrus"
)

func Serve(ctx context.Context, broker broker.MessageBroker) {
	dataChan := broker.StartGetData(ctx)

	for {
		select {
		case d := <-dataChan:
			err := broker.Commit(ctx, d.MessageUuid)
			if err != nil {
				logrus.Errorf("Failed to commit the messsage with UUID: %v: %v", d.MessageUuid.String(), err)
			}
		case <-ctx.Done():
			broker.Stop()
			logrus.Info("The application was stopped")
		}
	}
}
