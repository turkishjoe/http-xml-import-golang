package api

import (
	"context"
	"fmt"
	"net/http"
)

type ApiService struct {
}

func NewService() Service {
	return &ApiService{}
}

func (w *ApiService) Update(ctx context.Context) {
	url := "https://www.treasury.gov/ofac/downloads/sdn.xml"

	client := http.Client{
		Timeout: 3,
	}
	// Get the data
	resp, err := client.Get(url)
	if err != nil {
		fmt.Println("fuck")
		return
	}
	fmt.Printf("test")
	defer resp.Body.Close()
}
