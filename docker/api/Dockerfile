# syntax=docker/dockerfile:1

FROM golang:1.19

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY .env .env
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/engine/reference/builder/#copy
COPY internal ./internal
COPY cmd/api/api.go ./cmd/api/api.go

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /api cmd/api/api.go

EXPOSE 8080

# Run
CMD ["/api"]
