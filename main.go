package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/yayuyokitano/eggshellver/lib/cachecreator"
	followendpoint "github.com/yayuyokitano/eggshellver/lib/endpoints/follow"
	likeendpoint "github.com/yayuyokitano/eggshellver/lib/endpoints/like"
	playlistendpoint "github.com/yayuyokitano/eggshellver/lib/endpoints/playlist"
	"github.com/yayuyokitano/eggshellver/lib/endpoints/timeline"
	userendpoint "github.com/yayuyokitano/eggshellver/lib/endpoints/user"
	userstubendpoint "github.com/yayuyokitano/eggshellver/lib/endpoints/userstub"
	wsendpoint "github.com/yayuyokitano/eggshellver/lib/endpoints/ws"
	"github.com/yayuyokitano/eggshellver/lib/hub"
	"github.com/yayuyokitano/eggshellver/lib/logging"
	"github.com/yayuyokitano/eggshellver/lib/router"
	"github.com/yayuyokitano/eggshellver/lib/services"
)

func main() {
	switch os.Args[1] {
	case "migrate":
		fmt.Println("Performing migration...")
		performMigration(true)
		fmt.Println("Migration complete!")
		return
	case "createcache":
		services.Start()
		defer services.Stop()
		go logging.ServeLogs()
		cachecreator.AttemptRunPartialCache()
		fmt.Println("Cache creation complete!")
		return
	case "start":
		fmt.Println("Starting server...")
	default:
		fmt.Println("Invalid command")
		return
	}
	services.Start()
	hub.Init()
	defer services.Stop()
	fmt.Println("Connected to Postgres!")

	startServer()
}

func startServer() {
	router.Handle("/follows", router.Methods{
		POST:   followendpoint.Post,
		GET:    followendpoint.Get,
		PUT:    followendpoint.Put,
		DELETE: router.ReturnMethodNotAllowed,
	})
	router.Handle("/follow/", router.Methods{
		POST:   followendpoint.Toggle,
		GET:    router.ReturnMethodNotAllowed,
		PUT:    router.ReturnMethodNotAllowed,
		DELETE: router.ReturnMethodNotAllowed,
	})
	router.Handle("/likes", router.Methods{
		POST:   likeendpoint.Post,
		GET:    likeendpoint.Get,
		PUT:    likeendpoint.Put,
		DELETE: router.ReturnMethodNotAllowed,
	})
	router.Handle("/like/", router.Methods{
		POST:   likeendpoint.Toggle,
		GET:    router.ReturnMethodNotAllowed,
		PUT:    router.ReturnMethodNotAllowed,
		DELETE: router.ReturnMethodNotAllowed,
	})
	router.Handle("/playlists", router.Methods{
		POST:   playlistendpoint.Post,
		GET:    playlistendpoint.Get,
		PUT:    playlistendpoint.Put,
		DELETE: playlistendpoint.Delete,
	})
	router.Handle("/users", router.Methods{
		POST:   userendpoint.Post,
		GET:    userendpoint.Get,
		PUT:    router.ReturnMethodNotAllowed,
		DELETE: userendpoint.Delete,
	})
	router.Handle("/twitterauth", router.Methods{
		POST:   userendpoint.PostTwitter,
		GET:    router.ReturnMethodNotAllowed,
		PUT:    router.ReturnMethodNotAllowed,
		DELETE: router.ReturnMethodNotAllowed,
	})
	router.Handle("/userstubs", router.Methods{
		POST:   userstubendpoint.Post,
		GET:    router.ReturnMethodNotAllowed,
		PUT:    router.ReturnMethodNotAllowed,
		DELETE: router.ReturnMethodNotAllowed,
	})
	router.Handle("/timeline", router.Methods{
		POST:   router.ReturnMethodNotAllowed,
		GET:    timeline.Get,
		PUT:    router.ReturnMethodNotAllowed,
		DELETE: router.ReturnMethodNotAllowed,
	})

	router.HandleWebsocket("/ws/join/", wsendpoint.Establish)
	router.HandleWebsocket("/ws/create/", wsendpoint.Create)
	router.Handle("/ws/list", router.Methods{
		POST:   router.ReturnMethodNotAllowed,
		GET:    wsendpoint.GetHubs,
		PUT:    router.ReturnMethodNotAllowed,
		DELETE: router.ReturnMethodNotAllowed,
	})

	go logging.ServeLogs()
	go cachecreator.StartCacheLoop(1 * time.Hour)
	fmt.Println("===========")
	fmt.Println("eggshellver v0.1.0")
	http.ListenAndServe(":10000", nil)
}

func performMigration(firstTime bool) {
	migrations := &migrate.FileMigrationSource{
		Dir: "./migrations",
	}

	db, err := sql.Open("pgx", fmt.Sprintf("postgresql://%s:%s@db:5432/%s", os.Getenv("POSTGRES_USER"), url.QueryEscape(os.Getenv("POSTGRES_PASSWORD")), os.Getenv("POSTGRES_DB")))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", os.Getenv("POSTGRES_DB")))
	if err != nil {
		fmt.Println("Failed to create database, probably already exists.")
	}

	_, err = db.Exec(fmt.Sprintf("CREATE USER %s WITH PASSWORD '%s'", os.Getenv("POSTGRES_GRAFANA_USER"), os.Getenv("POSTGRES_GRAFANA_PASSWORD")))
	if err != nil {
		fmt.Println("Failed to create user, probably already exists.")
	}

	_, err = db.Exec(fmt.Sprintf("GRANT pg_read_all_data TO %s", os.Getenv("POSTGRES_GRAFANA_USER")))
	if err != nil {
		fmt.Println("Failed to grant user permissions, probably already exists.")
	}

	n, err := migrate.Exec(db, "postgres", migrations, migrate.Up)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Applied %d migrations!\n", n)

}
