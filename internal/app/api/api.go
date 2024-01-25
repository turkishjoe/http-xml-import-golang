package api

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/turkishjoe/xml-parser/internal/app/api/domain"
	"github.com/turkishjoe/xml-parser/internal/app/api/repo"
	"github.com/turkishjoe/xml-parser/internal/pkg/individuals"
	"github.com/turkishjoe/xml-parser/internal/pkg/state"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

const BATCH_SIZE = 50
const SAVE_GOROUTINES = 2
const SDN_URL = "https://www.treasury.gov/ofac/downloads/sdn.xml"

type ApiService struct {
	individualRepo repo.IndividualRepo
	Logger         *log.Logger
	notifier       state.Notifier
}

func NewService(conn *pgxpool.Pool, logger *log.Logger) Service {
	return &ApiService{
		individualRepo: repo.CreateRepo(conn),
		Logger:         logger,
		notifier:       state.NewNotifier(),
	}
}

func (apiService *ApiService) Update(ctx context.Context) error {
	if apiService.notifier.ReadValue() {
		return errors.New("Service is processing import now")
	}

	go apiService.processUpdate(ctx)

	return nil
}

func (apiService *ApiService) processUpdate(ctx context.Context) {
	apiService.notifier.Notify(true)
	defer apiService.notifier.Notify(false)

	req, _ := http.NewRequest("GET", SDN_URL, nil)
	resp, _ := http.DefaultClient.Do(req)

	parserChannel := make(chan map[string]string, SAVE_GOROUTINES)
	parserInstance := individuals.NewParser(apiService.Logger)

	go parserInstance.Parse(resp.Body, parserChannel)
	wg := sync.WaitGroup{}

	for i := 0; i < SAVE_GOROUTINES; i++ {
		go apiService.parse(parserChannel, &wg)
		wg.Add(1)
	}

	wg.Wait()
}

func (apiService *ApiService) parse(parserChannel chan map[string]string, wg *sync.WaitGroup) {
	defer wg.Done()
	databaseMapper := map[string]string{
		"uid":       "id",
		"firstName": "first_name",
		"lastName":  "last_name",
	}

	batchData := make([]map[string]string, BATCH_SIZE)
	var countInCurrentBatch int = 0
	initialState := apiService.State(context.Background())

	for {
		parsedRow, ok := <-parserChannel

		if !ok {
			break
		}

		if initialState == Empty {
			for k, v := range parsedRow {
				parsedRow[databaseMapper[k]] = v
			}

			batchData = append(batchData, parsedRow)
			countInCurrentBatch++

			if countInCurrentBatch != BATCH_SIZE {
				continue
			}

			countInCurrentBatch = 0
			apiService.individualRepo.BatchInsert(batchData)
		}

		id, castingError := strconv.ParseInt(parsedRow["uid"], 10, 64)

		if castingError != nil {
			apiService.Logger.Printf("Cast error, sdn id is not integer:%s\n", parsedRow["id"])

			continue
		}

		databaseError := apiService.individualRepo.UpdateOrInsert(
			id,
			parsedRow["firstName"],
			parsedRow["lastName"],
		)

		//Возможно при таких случаях следует сразу возвращать 500-ку
		//Но также возможны варианты, когда сбои произошли только на части на данных
		//Что делать я бы решал в индивидуальном плане, для простоты логирую
		if databaseError != nil {
			apiService.Logger.Println("database_error:", databaseError)
		}
	}

	if countInCurrentBatch > 0 {
		apiService.individualRepo.BatchInsert(batchData)
	}
}

// В примерах не до конца понял поведение, которое должно быть. Увидел, что если в name указано только
// одно слово, то нашлось по last_name(в моем случае, я при таком раскладе ищу либо совпадение по first_name, либо
// по last_name, так как в ТЗ не обговоренно иного). Также если передать строку с пробелами, я беру в качестве
// first_name - первое слово, а в качестве last_name все остальные слова. В бд видел, что lastname не всегда
// представляет одно слово, потом процесс реален. Также у меня происходит обмен. (То есть last_name и first_name меняются
// и возвращаются все совпдаения, подробнее cм репозиторий). Сделано так, так как по ответам, не совсем понятно, как правильно
func (apiService *ApiService) GetNames(ctx context.Context, name string, searchType domain.SearchType) []domain.Individual {
	words := strings.Fields(name)

	if len(words) == 1 {
		res, err := apiService.individualRepo.GetNamesBySingleString(name, searchType)

		if err != nil {
			apiService.Logger.Println("Database error:", err)
		}

		return res
	}

	res, err := apiService.individualRepo.GetNamesByFirstAndLastName(
		words[0],
		strings.Join(words[1:], " "),
		searchType,
	)

	if err != nil {
		apiService.Logger.Println("Database error:", err)
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
