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

const insertNewRefComm = `INSERT INTO file_references (number, reference) VALUES ($1, $2)`

func (st PgReferenceStorage) Insert(ctx context.Context, num int64, ref string) error {
	_, err := st.connection.Exec(ctx, insertNewRefComm, num, ref)
	if err != nil {
		return fmt.Errorf("failed to insert data into the storage: %w", err)
	}

	return nil
}

const getRefCom = `SELECT reference FROM file_references WHERE number = $1`

func (st PgReferenceStorage) GetReference(ctx context.Context, num int64) (string, error) {
	row := st.connection.QueryRow(ctx, getRefCom, num)

	var ref string

	err := row.Scan(&ref)
	if err != nil {
		return "", fmt.Errorf("failed to read a storage row: %w", err)
	}

	return ref, nil
}

const getAllRefCom = `SELECT number, reference FROM file_references ORDER BY number`

func (st PgReferenceStorage) GetAllReferences(ctx context.Context) ([]string, error) {
	rows, err := st.connection.Query(ctx, getAllRefCom)
	if err != nil {
		return nil, fmt.Errorf("failed to read storage rows: %w", err)
	}

	defer rows.Close()

	refs := make([]string, 0)
	var num int64
	var ref string

	for rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("failed to process a row from storage: %w", err)
		}

		err := rows.Scan(&num, &ref)
		if err != nil {
			return nil, fmt.Errorf("failed to parse data from the storage: %w", err)
		}

		refs = append(refs, fmt.Sprintf("%v. %v", num, ref))
	}

	return refs, nil
}

const delRefCom = `DELETE FROM file_references WHERE num = $1`

func (st PgReferenceStorage) DeleteReference(ctx context.Context, num int64) error {
	_, err := st.connection.Exec(ctx, delRefCom, num)
	if err != nil {
		return fmt.Errorf("failed to delete storage row with the number %v: %w", num, err)
	}

	return nil
}
