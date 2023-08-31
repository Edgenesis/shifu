package template

import (
	"context"
)

type DBDriver interface {
	ConnectToDB(ctx context.Context) error
	InsertDataToDB(ctx context.Context, deviceName string, rawData []byte) error
	SendToDB(ctx context.Context, deviceName string, rawData []byte) error
}
