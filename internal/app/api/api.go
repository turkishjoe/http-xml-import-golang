package api

import (
	"context"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/turkishjoe/xml-parser/internal/pkg/individuals"
	"github.com/turkishjoe/xml-parser/internal/pkg/state"
	"net/http"
	"strings"
)

const SDN_URL = "https://www.treasury.gov/ofac/downloads/sdn.xml"

type ApiService struct {
	DatabaseConnection *pgxpool.Pool
	Logger             log.Logger
	Notifier           state.Notifier
}

func NewService(conn *pgxpool.Pool, logger log.Logger) Service {
	return &ApiService{
		DatabaseConnection: conn,
		Logger:             logger,
		Notifier:           state.NewNotifier(),
	}
}

func (apiService *ApiService) Update(ctx context.Context) {
	apiService.Notifier.Notify(true)
	defer apiService.Notifier.Notify(false)

	databaseMapper := map[string]string{
		"uid":       "id",
		"firstName": "first_name",
		"lastName":  "last_name",
	}

	req, _ := http.NewRequest("GET", SDN_URL, nil)
	resp, _ := http.DefaultClient.Do(req)

	parserChannel := make(chan map[string]string)
	parserInstance := individuals.NewParser(apiService.Logger)

	go parserInstance.Parse(resp.Body, parserChannel)

	for {
		parsedRow, ok := <-parserChannel

		if !ok {
			break
		}

		args := pgx.NamedArgs{}
		for k, v := range parsedRow {
			args[databaseMapper[k]] = v
		}

		query := "INSERT INTO individuals(id, first_name, last_name) " +
			"VALUES(@id, @first_name, @last_name) " +
			"ON CONFLICT(\"id\") DO UPDATE SET" +
			" first_name=EXCLUDED.first_name, last_name=EXCLUDED.last_name " +
			"RETURNING id"

		_, databaseErr := apiService.DatabaseConnection.Exec(
			context.Background(),
			query,
			args,
		)

		if databaseErr != nil {
			apiService.Logger.Log("database", databaseErr)
			panic(databaseErr)
		}
	}
}

func (apiService *ApiService) GetNames(ctx context.Context, name string, searchType SearchType) []Individual {
	var result []Individual
	var queryStringBuilder strings.Builder
	queryStringBuilder.WriteString("SELECT id, first_name, last_name from individuals WHERE ")

	likeString := "CONCAT('%',LOWER(@name),'%')"

	if searchType == Weak || searchType == Both {
		queryStringBuilder.WriteString(
			fmt.Sprintf(
				"LOWER(first_name) LIKE %s OR LOWER(last_name) LIKE %s",
				likeString,
				likeString,
			),
		)
	}

	if searchType == Strong || searchType == Both {
		if searchType == Both {
			queryStringBuilder.WriteString(" OR ")
		}

		queryStringBuilder.WriteString(
			fmt.Sprintf(
				"LOWER(first_name) = %s OR last_name = %s",
				likeString,
				likeString,
			),
		)
	}

	rows, err := apiService.DatabaseConnection.Query(context.Background(), queryStringBuilder.String(), pgx.NamedArgs{
		"name": name,
	})

	if err != nil {
		apiService.Logger.Log("database", err)
		return result
	}

	for rows.Next() {
		individual := Individual{}
		err = rows.Scan(&individual.Uid, &individual.FirstName, &individual.LastName)
		if err != nil {
			apiService.Logger.Log("scan_rows", err)
			return result
		}

		result = append(result, individual)
	}

	return result
}

func (apiService *ApiService) State(ctx context.Context) State {
	var res int64

	if apiService.Notifier.ReadValue() {
		return Updating
	}

	err := apiService.DatabaseConnection.QueryRow(context.Background(), "select id from individuals limit 1").Scan(&res)

	if err != nil {
		return Empty
	}

	return Ok
}
