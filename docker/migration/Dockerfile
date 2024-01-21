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
COPY migrations ./migrations
COPY ./cmd/migrate/main.go ./cmd/migrate/main.go
COPY ./migrations ./migrations
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/migrate cmd/migrate/main.go
ENTRYPOINT ["./migrate"]
CMD ["up"]
