package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"os"

	_ "github.com/jackc/pgx/v4/stdlib"
	migrate "github.com/rubenv/sql-migrate"
	followendpoint "github.com/yayuyokitano/eggshellver/lib/endpoints/follow"
	likeendpoint "github.com/yayuyokitano/eggshellver/lib/endpoints/like"
	playlistendpoint "github.com/yayuyokitano/eggshellver/lib/endpoints/playlist"
	userendpoint "github.com/yayuyokitano/eggshellver/lib/endpoints/user"
	userstubendpoint "github.com/yayuyokitano/eggshellver/lib/endpoints/userstub"
	"github.com/yayuyokitano/eggshellver/lib/router"
	"github.com/yayuyokitano/eggshellver/lib/services"
)

func main() {
	if os.Args[1] == "migrate" {
		fmt.Println("Performing migration...")
		performMigration(true)
		fmt.Println("Migration complete!")
		return
	}
	if os.Args[1] != "start" {
		return
	}
	fmt.Println("Starting server...")
	services.Start()
	defer services.Stop()
	fmt.Println("Connected to Postgres!")
	fmt.Println("===========")
	fmt.Println("eggshellver v0.1.0")

	startServer()
}

func startServer() {
	router.Handle("/follows", router.Methods{
		POST:   followendpoint.Post,
		GET:    followendpoint.Get,
		PUT:    followendpoint.Put,
		DELETE: followendpoint.Delete,
	})
	router.Handle("/likes", router.Methods{
		POST:   likeendpoint.Post,
		GET:    likeendpoint.Get,
		PUT:    likeendpoint.Put,
		DELETE: likeendpoint.Delete,
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
		DELETE: router.ReturnMethodNotAllowed,
	})
	router.Handle("/userstubs", router.Methods{
		POST:   userstubendpoint.Post,
		GET:    router.ReturnMethodNotAllowed,
		PUT:    router.ReturnMethodNotAllowed,
		DELETE: router.ReturnMethodNotAllowed,
	})
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

	n, err := migrate.Exec(db, "postgres", migrations, migrate.Up)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Applied %d migrations!\n", n)
}
