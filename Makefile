migrate-make:
	go run cmd/migrate/main.go create $(name) sql
migrate-help:
	go run cmd/migrate/main.go
migrate-up:
	go run cmd/migrate/main.go up
migrate-down:
	go run cmd/migrate/main.go down

