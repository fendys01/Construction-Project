package psql

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Connect
func Connect(dsn string) (*pgxpool.Pool, error) {
	// urlExample := "postgres://username:password@localhost:5432/database_name"
	conn, err := pgxpool.Connect(context.Background(), dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		return conn, err
	}

	return conn, nil
}

// Parsing Error
func ParseErr(err error) string {
	switch pqe := err.(type) {
	case *pgconn.PgError:

		// error duplicate
		if pqe.Code == "23505" {
			m := strings.ReplaceAll(pqe.Detail, "(", "")
			m = strings.ReplaceAll(m, ")", "")
			m = strings.ReplaceAll(m, "=", " = ")
			return m
		}
	}
	return err.Error()
}
