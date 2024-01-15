package lib

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
)

func OpenDbConnection() (*sql.DB, error) {
	db, err := sql.Open(
		os.Getenv("PSQL_DATABASE_DRIVE_NAME"),
		fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			os.Getenv("PSQL_HOST"),
			os.Getenv("PSQL_PORT"),
			os.Getenv("PSQL_USER"),
			os.Getenv("PSQL_PASS"),
			os.Getenv("PSQL_DBNAME"),
			os.Getenv("PSQL_SSLMODE"),
		),
	)
	if err != nil {
		Log(LOG_ERROR, "Opening Database Connection", fmt.Sprintf("Could not connect to database | reason: %v", err))
		return nil, errors.New("could not connect to database")
	}
	return db, nil
}
