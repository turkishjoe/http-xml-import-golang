package main

import (
	"context"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/turkishjoe/xml-parser/internal/app/api"
	"github.com/turkishjoe/xml-parser/internal/app/api/endpoints"
	"github.com/turkishjoe/xml-parser/internal/app/api/transport"
	"net"
	"net/http"
	"os"
)

const defaultHTTPHost = "localhost"
const defaultHTTPPort = "8081"

func main() {
	var logger log.Logger

	err := godotenv.Load()
	if err != nil {
		logger.Log("Fatal load env")
		return
	}

	// urlExample := "postgres://username:password@localhost:5432/database_name"
	postgresUrl := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_DATABASE"),
	)

	conn, err := pgxpool.New(context.Background(), postgresUrl)
	if err != nil {
		logger.Log("Unable to connect to database:", err)
		os.Exit(1)
	}
	defer conn.Close()

	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	httpAddr := net.JoinHostPort(envString("HTTP_HOST", defaultHTTPHost), envString("HTTP_PORT", defaultHTTPPort))

	var (
		service     = api.NewService(conn, &logger)
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

func envString(env, fallback string) string {
	e := os.Getenv(env)
	if e == "" {
		return fallback
	}
	return e
}
