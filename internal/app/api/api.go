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
	"net/http"
)

const (
	BUFFER_SIZE         = 32 * 1024
	SND_INDIVIDIAL_TYPE = "Individual"
	SDN_URL             = "https://www.treasury.gov/ofac/downloads/sdn.xml"
)

type ApiService struct {
	DatabaseConnection *pgxpool.Pool
	Logger             log.Logger
}

func NewService(conn *pgxpool.Pool, logger log.Logger) Service {
	return &ApiService{
		DatabaseConnection: conn,
		Logger:             logger,
	}
}

func (apiService *ApiService) Update(ctx context.Context) {
	req, _ := http.NewRequest("GET", SDN_URL, nil)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	buf := bufio.NewReaderSize(resp.Body, BUFFER_SIZE)
	parser := xmlparser.NewXMLParser(buf, "sdnEntry")

	requiredFields := []string{}
	optinalFields := []string{"firstName", "lastName", "title"}
	databaseMapper := map[string]string{
		"firstName": "first_name",
		"lastName":  "last_name",
		"title":     "title",
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

		query := "INSERT INTO individuals(id, first_name, last_name, title) " +
			"VALUES(@id, @first_name, @last_name, @title) "

		_, databaseErr := apiService.DatabaseConnection.Exec(
			context.Background(),
			query,
			args,
		)

		if databaseErr != nil {
			apiService.Logger.Log("parse", databaseErr)
			panic(databaseErr)
		}
	}
}
