package postgresql

import (
	"database/sql"

	_ "github.com/lib/pq"
)

func SetupPostgresConnection(url string) (*sql.DB, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}

	// TODO: Configure connection pool

	return db, nil
}
