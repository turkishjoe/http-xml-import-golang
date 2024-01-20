package individuals

import (
	"bufio"
	"errors"
	"github.com/go-kit/kit/log"
	xmlparser "github.com/tamerh/xml-stream-parser"
	"io"
	"sync"
)

const (
	BUFFER_SIZE         = 32 * 1024
	SND_INDIVIDIAL_TYPE = "Individual"
	PARSE_GOROUTINE     = 2
)

var requiredFields = []string{}
var optionalFields = []string{"firstName", "lastName"}

type Parser struct {
	logger log.Logger
}

func NewParser(log log.Logger) *Parser {
	return &Parser{
		logger: log,
	}
}

func (parser *Parser) Parse(input io.ReadCloser, output chan map[string]string) {
	defer input.Close()
	buf := bufio.NewReaderSize(input, BUFFER_SIZE)
	xmlParser := xmlparser.NewXMLParser(buf, "sdnEntry")
	inputXmlChan := make(chan *xmlparser.XMLElement, PARSE_GOROUTINE)
	wg := sync.WaitGroup{}

	for i := 0; i < PARSE_GOROUTINE; i++ {
		go parser.parseGoroutineInit(inputXmlChan, output, &wg)
		wg.Add(1)
	}

	for xml := range xmlParser.Stream() {
		inputXmlChan <- xml
	}

	wg.Wait()
	close(output)
}

func (parser *Parser) parseGoroutineInit(input chan *xmlparser.XMLElement, output chan map[string]string, wg *sync.WaitGroup) {
	for {
		xml, ok := <-input

		if !ok {
			break
		}

		res, err := parser.parseItem(xml)

		if err != nil {
			parser.logger.Log("parse", err)
			continue
		}

		output <- res
	}

	wg.Done()
}

func (parser *Parser) parseItem(xml *xmlparser.XMLElement) (map[string]string, error) {
	args := map[string]string{}

	uidElement, hasUid := xml.Childs["uid"]
	uid := uidElement[0].InnerText
	args["uid"] = uid

	//id обрабатываем отдельно, чтобы в случае дальнеших ошибок, писать id записи
	if !hasUid {
		return nil, errors.New("Uuid is not set, move to next iteration")
	}

	sdnType, hasSdnType := xml.Childs["sdnType"]

	if !hasSdnType {
		return nil, errors.New("Missing sdnType" + " id:" + uid)
	}

	if sdnType[0].InnerText != SND_INDIVIDIAL_TYPE {
		return nil, errors.New("Sdn type not individual " + "id:" + uid)
	}

	var requiredFieldError error
	for _, requiredField := range requiredFields {
		value, hasField := xml.Childs[requiredField]
		if !hasField {
			requiredFieldError = errors.New("Failed to parse required field:" + requiredField)
			break
		}

		args[requiredField] = value[0].InnerText
	}

	if requiredFieldError != nil {
		return nil, requiredFieldError
	}

	for _, optionalField := range optionalFields {
		value, hasField := xml.Childs[optionalField]

		if !hasField {
			continue
		}

		args[optionalField] = value[0].InnerText
	}

	return args, nil
}
