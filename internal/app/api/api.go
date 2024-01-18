package api

import (
	"bufio"
	"context"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/jackc/pgx/v5"
	"github.com/tamerh/xml-stream-parser"
	"net/http"
)

const (
	BUFFER_SIZE = 32 * 1024
	SDN_URL     = "https://www.treasury.gov/ofac/downloads/sdn.xml"
)

type ApiService struct {
	DatabaseConnection *pgx.Conn
	Logger             *log.Logger
}

func NewService(conn *pgx.Conn, logger *log.Logger) Service {
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

	for xml := range parser.Stream() {
		uidElement, hasUid := xml.Childs["uid"]
		firstElement, hasFirstName := xml.Childs["first_name"]
		lastElement, hasLastName := xml.Childs["last_name"]
		titleElement, hasTitle := xml.Childs["title"]

		if !hasUid {
			apiService.Logger.Log("Uuid is not set")

			continue
		}

		apiService.DatabaseConnection.Query(
			context.Background(),
			"INSERT INTO VALUES($1, $2, $3, $4) ",
			uidElement[0].InnerText,
			firstElement[0].InnerText,
			lastElement[0].InnerText,
			titleElement[0].InnerText,
		)

		//fmt.Println(el[0].InnerText)
	}
}
