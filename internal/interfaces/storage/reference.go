package storage

import "context"

type ReferenceStorage interface {
	Insert(ctx context.Context, num int64, ref string) error
	GetReference(ctx context.Context, num int64) (string, error)
	GetAllReferences(ctx context.Context) ([]string, error)
}
