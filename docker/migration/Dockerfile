# syntax=docker/dockerfile:1

FROM golang:1.19

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY ../../.env .env
COPY ../../go.mod go.sum ./
RUN go mod download

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/engine/reference/builder/#copy
COPY ../../migrations ./migration
COPY ../../cmd/api/migration.go ./cmd/api/migration.go

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /app
CMD ["go run cmd/migrate/main.go up"]
