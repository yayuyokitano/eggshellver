package services

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
)

var Pool *pgxpool.Pool
var IsTesting bool

func Start() (err error) {
	if os.Getenv("TESTING") == "true" {
		IsTesting = true
	}

	connectionString := fmt.Sprintf("postgresql://%s:%s@db:5432/%s?pool_max_conns=100",
		os.Getenv("POSTGRES_USER"), url.QueryEscape(os.Getenv("POSTGRES_PASSWORD")), os.Getenv("POSTGRES_DB"))
	Pool, err = pgxpool.Connect(context.Background(), connectionString)
	if err != nil {
		return
	}
	return
}

func Stop() {
	if Pool != nil {
		Pool.Close()
	}
}
