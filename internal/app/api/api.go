package api

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tamerh/xml-stream-parser"
	"github.com/turkishjoe/xml-parser/internal/pkg/state"
	"net/http"
	"strings"
)

const (
	BUFFER_SIZE         = 32 * 1024
	SND_INDIVIDIAL_TYPE = "Individual"
	SDN_URL             = "https://www.treasury.gov/ofac/downloads/sdn.xml"
)

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
	req, _ := http.NewRequest("GET", SDN_URL, nil)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	buf := bufio.NewReaderSize(resp.Body, BUFFER_SIZE)
	parser := xmlparser.NewXMLParser(buf, "sdnEntry")

	requiredFields := []string{}
	optinalFields := []string{"firstName", "lastName"}
	databaseMapper := map[string]string{
		"firstName": "first_name",
		"lastName":  "last_name",
	}

	for xml := range parser.Stream() {
		args := pgx.NamedArgs{}

		uidElement, hasUid := xml.Childs["uid"]
		uid := uidElement[0].InnerText
		args["id"] = uid

		//id обрабатываем отдельно, чтобы в случае дальнеших ошибок, писать id записи
		if !hasUid {
			apiService.Logger.Log("parse", "Uuid is not set, move to next iteration")

			continue
		}

		sdnType, hasSdnType := xml.Childs["sdnType"]

		if !hasSdnType {
			apiService.Logger.Log("parse", "Missing sdnType", "id", uid)

			continue
		}

		if sdnType[0].InnerText != SND_INDIVIDIAL_TYPE {
			apiService.Logger.Log("parse", "Sdn type not individual", "id", uid)

			continue
		}

		var requiredFieldError error
		for _, requiredField := range requiredFields {
			value, hasField := xml.Childs[requiredField]
			if !hasField {
				requiredFieldError = errors.New("Failed to parse required field:" + requiredField)
				break
			}

			args[databaseMapper[requiredField]] = value[0].InnerText
		}

		if requiredFieldError != nil {
			apiService.Logger.Log("parse", requiredFieldError)
			continue
		}

		for _, optionalField := range optinalFields {
			value, hasField := xml.Childs[optionalField]

			if !hasField {
				apiService.Logger.Log("parse", "Failed to parse optional field:"+optionalField,
					"id:", uid,
				)
				break
			}
			args[databaseMapper[optionalField]] = value[0].InnerText
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
