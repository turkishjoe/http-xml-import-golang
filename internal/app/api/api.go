package api

import (
	"context"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/turkishjoe/xml-parser/internal/app/api/repo"
	"github.com/turkishjoe/xml-parser/internal/pkg/individuals"
	"github.com/turkishjoe/xml-parser/internal/pkg/state"
	"net/http"
	"strings"
	"sync"
	"time"
)

const BATCH_SIZE = 100
const SAVE_GOROUTINES = 3
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

	conflictString := "ON CONFLICT(\"id\") DO UPDATE SET" +
		" first_name=EXCLUDED.first_name, last_name=EXCLUDED.last_name " +
		"RETURNING id"

	query := "INSERT INTO individuals(id, first_name, last_name) " +
		"VALUES($1, $2, $3) " +
		conflictString

	batchData := make([]map[string]string, BATCH_SIZE)
	var i int = 0

	for {
		parsedRow, ok := <-parserChannel

		if !ok {
			break
		}

		for k, v := range parsedRow {
			parsedRow[databaseMapper[k]] = v
		}

		batchData = append(batchData, parsedRow)
		i++

		if i != BATCH_SIZE {
			continue
		}

		batch := pgx.Batch{}

		for _, v := range batchData {
			batch.Queue(query, v["id"], v["first_name"], v["last_name"])
		}

		apiService.databaseConnection.SendBatch(
			context.Background(),
			&batch,
		)

		/*if databaseErr != nil {
			apiService.Logger.Log("database", databaseErr)
			panic(databaseErr)
		}*/
	}

	if i > 0 {
		batch := pgx.Batch{}
		for _, v := range batchData {
			batch.Queue(query, v["id"], v["first_name"], v["last_name"])
		}

		bc := apiService.databaseConnection.SendBatch(
			context.Background(),
			&batch,
		)

		bc.Close()
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

	rows, err := apiService.databaseConnection.Query(context.Background(), queryStringBuilder.String(), pgx.NamedArgs{
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
	if apiService.notifier.ReadValue() {
		return Updating
	}

	var res int64
	err := apiService.databaseConnection.QueryRow(context.Background(), "select id from individuals limit 1").Scan(&res)

	if err != nil {
		return Empty
	}

	return Ok
}
