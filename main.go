package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	ctx, cancel := context.WithCancel(context.Background())

	dbUri, dbName := os.Getenv("MONGO_URI"), os.Getenv("MONGO_DB_NAME")
	db, closeDB := OpenMongoDBConn(ctx, dbUri, dbName)
	defer closeDB()

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "1234"
	}

	httpAddr := ":" + httpPort

	jobChn := make(chan interface{}, 1000)
	defer close(jobChn)

	hdl, closeHdl := NewHandler(ctx, db)
	defer closeHdl()

	errChn := make(chan error)

	go func() {
		log.Printf("http server running on port %s\n", httpAddr)
		errChn <- http.ListenAndServe(httpAddr, hdl)
	}()

	go func() {
		signalChn := make(chan os.Signal, 1)
		signal.Notify(signalChn, syscall.SIGINT, syscall.SIGTERM)
		errChn <- fmt.Errorf("%s", <-signalChn)
	}()

	log.Printf("server exit: %v\n", <-errChn)

	cancel()
}
