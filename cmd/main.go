package main

import (
	"context"
	app2 "github.com/DaniilOr/marketcap/cmd/app"
	marketcap2 "github.com/DaniilOr/marketcap/pkg/marketcap"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net"
	"net/http"
	"os"
)

const (
	defaultPort        = "9999"
	defaultHost        = "0.0.0.0"
	defaultMongoDB     = "kadex"
	defaultMongoDSN    = "mongodb://localhost:27017/" + defaultMongoDB
)


func main() {
	port, ok := os.LookupEnv("APP_PORT")
	if !ok {
		port = defaultPort
	}

	host, ok := os.LookupEnv("APP_HOST")
	if !ok {
		host = defaultHost
	}

	mongoDsn, ok := os.LookupEnv("Mongo_DSN")
	if !ok {
		mongoDsn = defaultMongoDSN
	}
	mongoDB, ok := os.LookupEnv("Mongo_DB")
	if !ok {
		mongoDB = defaultMongoDB
	}
	if err := execute(net.JoinHostPort(host, port), mongoDsn, mongoDB); err != nil {
		os.Exit(1)
	}
}

func execute(addr string, mongoDsn string, mongDB string) error {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoDsn))
	if err != nil {
		log.Print(err)
		return err
	}

	database := client.Database(mongDB)
	marketcapSvc := marketcap2.CreateService(database)
	router := chi.NewRouter()
	application := app2.NewServer(marketcapSvc, router)
	err = application.Init()
	if err != nil {
		log.Print(err)
		return err
	}

	server := &http.Server{
		Addr:    addr,
		Handler: application,
	}
	return server.ListenAndServe()
}

