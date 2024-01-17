package main

import (
	"flag"
	"fmt"
	//	"github.com/pressly/goose"
	"github.com/pressly/goose/v3"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

const (
	dialect = "pgx"
)

var (
	flags = flag.NewFlagSet("migrate", flag.ExitOnError)
)

func main() {
	var dir string

	err := godotenv.Load()
	if err != nil {
		return
	}

	fmt.Println(os.Getenv("MIGRATION_DIR"))

	if len(os.Getenv("MIGRATION_DIR")) > 0 {
		dir = os.Getenv("MIGRARTION_DIR")
	} else {
		dir = *flags.String("dir", "migrations", "directory with migration files")
	}

	flags.Usage = usage
	flags.Parse(os.Args[1:])

	//	dbString := "host=localhost user=myapp_user password=myapp_pass dbname=myapp_db port=5432 sslmode=disable"
	dbString := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_DATABASE"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_SSL_MODE"),
	)

	args := flags.Args()
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		flags.Usage()
		return
	}

	command := args[0]

	db, err := goose.OpenDBWithDriver(dialect, dbString)
	if err != nil {
		log.Fatalf(err.Error())
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf(err.Error())
		}
	}()

	if err := goose.Run(command, db, dir, args[1:]...); err != nil {
		log.Fatalf("migrate %v: %v", command, err)
	}
}

func usage() {
	fmt.Println(usagePrefix)
	flags.PrintDefaults()
	fmt.Println(usageCommands)
}

var (
	usagePrefix = `Usage: migrate COMMAND
Examples:
    migrate status
`

	usageCommands = `
Commands:
    up                   Migrate the DB to the most recent version available
    up-by-one            Migrate the DB up by 1
    up-to VERSION        Migrate the DB to a specific VERSION
    down                 Roll back the version by 1
    down-to VERSION      Roll back to a specific VERSION
    redo                 Re-run the latest migration
    reset                Roll back all migrations
    status               Dump the migration status for the current DB
    version              Print the current version of the database
    create NAME [sql|go] Creates new migration file with the current timestamp
    fix                  Apply sequential ordering to migrations`
)
