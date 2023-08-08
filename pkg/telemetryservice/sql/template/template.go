package template

import (
	"context"
)

type DBDriver interface {
	ConnectToDB(ctx context.Context) error
	InsertDataToDB(ctx context.Context, rawData []byte) error
}
