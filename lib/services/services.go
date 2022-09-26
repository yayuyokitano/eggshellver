package services

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/yayuyokitano/eggshellver/lib/hub"
)

var Pool *pgxpool.Pool
var IsTesting bool
var hubs map[string]*hub.AuthedHub

func Start() (err error) {
	if os.Getenv("TESTING") == "true" {
		IsTesting = true
	}

	hubs = make(map[string]*hub.AuthedHub)

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

func AttachHub(user string) {
	hubs[user] = &hub.AuthedHub{
		Hub:       hub.NewHub(),
		Owner:     user,
		Blocklist: make(map[string]bool),
	}
	go hubs[user].Hub.Run()
}

func GetHub(user string) *hub.AuthedHub {
	return hubs[user]
}
