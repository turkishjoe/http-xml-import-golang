package individuals

import (
	"io"
	"log"
)

type Parser interface {
	Parse(input io.ReadCloser, output chan map[string]string)
}

func NewParser(log *log.Logger) Parser {
	return &xmlStreamParser{
		logger: log,
	}
}
