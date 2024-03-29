package individuals

import (
	"bufio"
	"errors"
	xmlparser "github.com/tamerh/xml-stream-parser"
	"io"
	"log"
	"sync"
)

const (
	BUFFER_SIZE         = 32 * 1024
	SND_INDIVIDIAL_TYPE = "Individual"
	PARSE_GOROUTINE     = 4
)

var requiredFields = []string{}

// Считаем, что если firstName, lastName не пришли, то сохранится пустая строка
// Если такое поведение не устраиствает можно перенести в массив выше
var optionalFields = []string{"firstName", "lastName"}

type xmlStreamParser struct {
	logger *log.Logger
}

func (parser *xmlStreamParser) Parse(input io.ReadCloser, output chan map[string]string) {
	defer input.Close()
	buf := bufio.NewReaderSize(input, BUFFER_SIZE)
	xmlParser := xmlparser.NewXMLParser(buf, "sdnEntry")
	inputXmlChan := xmlParser.Stream()
	wg := sync.WaitGroup{}

	for i := 0; i < PARSE_GOROUTINE; i++ {
		wg.Add(1)
		go parser.parseGoroutineInit(inputXmlChan, output, &wg)
	}

	wg.Wait()
	close(output)
}

func (parser *xmlStreamParser) parseGoroutineInit(input chan *xmlparser.XMLElement, output chan map[string]string, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		xml, ok := <-input
		if !ok {
			break
		}

		res, err := parser.parseItem(xml)

		if err != nil {
			parser.logger.Println("parse_error:", err)
			continue
		}

		output <- res
	}
}

func (parser *xmlStreamParser) parseItem(xml *xmlparser.XMLElement) (map[string]string, error) {
	args := map[string]string{}

	uidElement, hasUid := xml.Childs["uid"]
	uid := uidElement[0].InnerText
	args["uid"] = uid

	//id обрабатываем отдельно, чтобы в случае дальнеших ошибок, писать id записи
	if !hasUid {
		return nil, errors.New("Uid is not set, move to next iteration")
	}

	sdnType, hasSdnType := xml.Childs["sdnType"]

	if !hasSdnType {
		return nil, errors.New("Missing sdnType id:" + uid)
	}

	if sdnType[0].InnerText != SND_INDIVIDIAL_TYPE {
		return nil, errors.New("Sdn type not individual id:" + uid)
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
