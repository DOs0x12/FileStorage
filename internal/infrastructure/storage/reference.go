package bd

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PgReferenceStorage struct {
	connection *pgxpool.Pool
}

const tableComm = `
CREATE TABLE IF NOT EXISTS file_references (
	number integer UNIQUE,
	reference varchar(4096) NOT NULL
)`

type StorageConf struct {
	Address,
	Database,
	User,
	Pass string
}

func NewPgRefStorage(ctx context.Context, conf StorageConf) (PgReferenceStorage, error) {
	maxLifeTime := "8760h"
	sslMode := "disable"
	connStr := fmt.Sprintf(
		"host=%v dbname=%v user=%v password=%v pool_max_conn_lifetime=%v sslmode=%v",
		conf.Address, conf.Database, conf.User, conf.Pass, maxLifeTime, sslMode)
	connConf, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return PgReferenceStorage{}, fmt.Errorf("the config for the database was not parsed: %v", err)
	}

	conn, err := pgxpool.NewWithConfig(ctx, connConf)
	if err != nil {
		return PgReferenceStorage{}, fmt.Errorf("failed to connect to the storage server: %w", err)
	}

	_, err = conn.Exec(ctx, tableComm)
	if err != nil {
		return PgReferenceStorage{}, fmt.Errorf("failed to create a table in the storage: %w", err)
	}

	return PgReferenceStorage{connection: conn}, nil
}

const insertNewRefComm = `INSERT INTO commands (number, reference) VALUES ($1, $2)`

func (st PgReferenceStorage) Insert(ctx context.Context, num int64, ref string) error {
	_, err := st.connection.Exec(ctx, insertNewRefComm, num, ref)
	if err != nil {
		return fmt.Errorf("failed to insert data into the storage: %w", err)
	}

	return nil
}
