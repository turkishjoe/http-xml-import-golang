package main

import (
	"github.com/go-kit/kit/log"
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
		return
	}

	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	httpAddr := net.JoinHostPort(envString("HTTP_HOST", defaultHTTPHost), envString("HTTP_PORT", defaultHTTPPort))

	var (
		service     = api.NewService()
		eps         = endpoints.NewEndpoints(service)
		httpHandler = transport.NewHTTPHandler(eps)
	)

	// The HTTP listener mounts the Go kit HTTP handler we created.
	httpListener, errListen := net.Listen("tcp", httpAddr)
	if errListen != nil {
		logger.Log("transport", "HTTP", "during", "Listen", "err", errListen)
		os.Exit(1)
	}

	err := http.Serve(httpListener, httpHandler)

	if err != nil {
		logger.Log("transport", "HTTP", "during", "Listen", "err", err)
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
