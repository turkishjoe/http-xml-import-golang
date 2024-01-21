package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/turkishjoe/xml-parser/internal/app/api"
	"github.com/turkishjoe/xml-parser/internal/app/api/endpoints"
	"github.com/turkishjoe/xml-parser/internal/app/api/transport"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
)

var logger *log.Logger

func main() {
	envError := godotenv.Load()
	if envError != nil {
		logger.Fatal("Fatal load env:", envError)
		return
	}

	logFileWriter, err := os.OpenFile(os.Getenv("LOG_FILE"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer logFileWriter.Close()

	logger = log.New(logFileWriter, "", 0644)

	conn := initDbPoll()
	defer conn.Close()

	httpAddr := net.JoinHostPort(os.Getenv("HTTP_HOST"), os.Getenv("HTTP_PORT"))

	var (
		service     = api.NewService(conn, logger)
		eps         = endpoints.NewEndpoints(service)
		httpHandler = transport.NewHTTPHandler(eps)
	)

	httpListener, errListen := net.Listen("tcp", httpAddr)
	if errListen != nil {
		logger.Println("Error http transport", errListen)
		os.Exit(1)
	}

	errService := http.Serve(httpListener, httpHandler)

	if errService != nil {
		logger.Println("Error http listen", errListen)
		os.Exit(1)
	}
}

func initDbPoll() *pgxpool.Pool {
	dbPort, parseError := strconv.Atoi(os.Getenv("DB_PORT"))

	if parseError != nil {
		logger.Fatal("Unable to read config:", parseError)
	}

	postgresUrl := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		dbPort,
		os.Getenv("DB_DATABASE"),
	)

	conn, err := pgxpool.New(context.Background(), postgresUrl)
	if err != nil {
		logger.Fatal("Unable to connect to database:", err)
	}

	return conn
}
