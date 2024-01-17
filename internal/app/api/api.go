package api

import (
	"bufio"
	"context"
	"fmt"
	"github.com/tamerh/xml-stream-parser"
	"net/http"
)

type ApiService struct {
}

func NewService() Service {
	return &ApiService{}
}

func (w *ApiService) Update(ctx context.Context) {
	fmt.Println("start")
	url := "https://www.treasury.gov/ofac/downloads/sdn.xml"
	req, _ := http.NewRequest("GET", url, nil)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	fmt.Println("Cool")
	buf := bufio.NewReaderSize(resp.Body, 32*1024)

	parser := xmlparser.NewXMLParser(buf, "sdnEntry")

	for xml := range parser.Stream() {
		el := xml.Childs["uid"]

		fmt.Println(el[0].InnerText)
	}
}
