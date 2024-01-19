package main

import (
	"context"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/turkishjoe/xml-parser/internal/app/api"
	"github.com/turkishjoe/xml-parser/internal/app/api/endpoints"
	"github.com/turkishjoe/xml-parser/internal/app/api/transport"
	"net"
	"net/http"
	"os"
	"strconv"
)

const defaultHTTPHost = "localhost"
const defaultHTTPPort = "8081"

var logger log.Logger

func main() {
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))

	err := godotenv.Load()
	if err != nil {
		logger.Log("Fatal load env")
		return
	}

	conn := initDbPoll()

	defer conn.Close()

	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	httpAddr := net.JoinHostPort(envString("HTTP_HOST", defaultHTTPHost), envString("HTTP_PORT", defaultHTTPPort))

	var (
		service     = api.NewService(conn, logger)
		eps         = endpoints.NewEndpoints(service)
		httpHandler = transport.NewHTTPHandler(eps)
	)

	// The HTTP listener mounts the Go kit HTTP handler we created.
	httpListener, errListen := net.Listen("tcp", httpAddr)
	if errListen != nil {
		logger.Log("transport", "HTTP", "during", "Listen", "err", errListen)
		os.Exit(1)
	}

	errService := http.Serve(httpListener, httpHandler)

	if errService != nil {
		logger.Log("transport", "HTTP", "during", "Listen", "err", errService)
		os.Exit(1)
	}
}

func initDbPoll() *pgxpool.Pool {
	dbPort, parseError := strconv.Atoi(os.Getenv("DB_PORT"))

	if parseError != nil {
		logger.Log("Unable to read config:", parseError)
		os.Exit(1)
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
		logger.Log("Unable to connect to database:", err)
		os.Exit(1)
	}

	return conn
}

func envString(env, fallback string) string {
	e := os.Getenv(env)
	if e == "" {
		return fallback
	}
	return e
}
