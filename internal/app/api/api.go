package api

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/turkishjoe/xml-parser/internal/app/api/domain"
	"github.com/turkishjoe/xml-parser/internal/app/api/repo"
	"github.com/turkishjoe/xml-parser/internal/pkg/individuals"
	"github.com/turkishjoe/xml-parser/internal/pkg/state"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const BATCH_SIZE = 100
const SAVE_GOROUTINES = 4
const SDN_URL = "https://www.treasury.gov/ofac/downloads/sdn.xml"

type ApiService struct {
	databaseConnection *pgxpool.Pool
	individualRepo     repo.IndividualRepo
	Logger             log.Logger
	notifier           state.Notifier
}

func NewService(conn *pgxpool.Pool, logger log.Logger) Service {
	return &ApiService{
		databaseConnection: conn,
		individualRepo:     repo.CreateRepo(conn),
		Logger:             logger,
		notifier:           state.NewNotifier(),
	}
}

func (apiService *ApiService) Update(ctx context.Context) {
	if apiService.notifier.ReadValue() {
		return
	}

	apiService.notifier.Notify(true)
	defer apiService.notifier.Notify(false)

	req, _ := http.NewRequest("GET", SDN_URL, nil)
	resp, _ := http.DefaultClient.Do(req)

	parserChannel := make(chan map[string]string, SAVE_GOROUTINES)
	parserInstance := individuals.NewParser(apiService.Logger)

	go parserInstance.Parse(resp.Body, parserChannel)
	start := time.Now()
	wg := sync.WaitGroup{}

	for i := 0; i < SAVE_GOROUTINES; i++ {
		go apiService.parse(parserChannel, &wg)
		wg.Add(1)
	}

	wg.Wait()
	apiService.Logger.Log("time_elapsed", time.Since(start))
}

func (apiService *ApiService) parse(parserChannel chan map[string]string, wg *sync.WaitGroup) {
	defer wg.Done()
	databaseMapper := map[string]string{
		"uid":       "id",
		"firstName": "first_name",
		"lastName":  "last_name",
	}

	batchData := make([]map[string]string, BATCH_SIZE)
	var i int = 0
	initialState := apiService.State(context.Background())

	for {
		parsedRow, ok := <-parserChannel

		if !ok {
			break
		}

		if initialState != Empty {
			id, parseError := strconv.ParseInt(parsedRow["uid"], 10, 64)

			if parseError != nil {
				apiService.Logger.Log("parse_error", "Id is not integer:", parsedRow["id"])

				continue
			}

			databaseError := apiService.individualRepo.UpdateOrInsert(
				id,
				parsedRow["firstName"],
				parsedRow["lastName"],
			)

			if databaseError != nil {
				apiService.Logger.Log("database_error", databaseError)

				continue
			}

			continue
		}

		for k, v := range parsedRow {
			parsedRow[databaseMapper[k]] = v
		}

		batchData = append(batchData, parsedRow)
		i++

		if i != BATCH_SIZE {
			continue
		}

		apiService.individualRepo.BatchInsert(batchData)
	}

	if i > 0 {
		apiService.individualRepo.BatchInsert(batchData)
	}
}

func (apiService *ApiService) GetNames(ctx context.Context, name string, searchType domain.SearchType) []domain.Individual {
	res, err := apiService.individualRepo.GetNames(name, searchType)

	if err != nil {
		apiService.Logger.Log("database_error", err)
	}

	return res
}

func (apiService *ApiService) State(ctx context.Context) State {
	if apiService.notifier.ReadValue() {
		return Updating
	}

	if apiService.individualRepo.IsEmpty() {
		return Empty
	}

	return Ok
}
