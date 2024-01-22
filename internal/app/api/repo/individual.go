package repo

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/turkishjoe/xml-parser/internal/app/api/domain"
	"strings"
)

// Можно вынести декларацию интерфейса
type IndividualRepo interface {
	IsEmpty() bool
	Insert(id int64, firstName string, lastName string) error
	UpdateOrInsert(id int64, firstName string, lastName string) error
	BatchInsert(batchData []map[string]string)
	GetNamesBySingleString(name string, searchType domain.SearchType) ([]domain.Individual, error)
	GetNamesByFirstAndLastName(firstName, lastName string, searchType domain.SearchType) ([]domain.Individual, error)
}

type individualRepoImp struct {
	databaseConnection *pgxpool.Pool
}

func CreateRepo(DatabaseConnection *pgxpool.Pool) IndividualRepo {
	return &individualRepoImp{
		databaseConnection: DatabaseConnection,
	}
}

func (individualRepo *individualRepoImp) IsEmpty() bool {
	var res int64
	err := individualRepo.databaseConnection.QueryRow(context.Background(), "select id from individuals limit 1").Scan(&res)

	return err != nil
}

func (individualRepo *individualRepoImp) Insert(id int64, firstName string, lastName string) error {
	query := `INSERT INTO individuals(id, first_name, last_name)
          VALUES($1, $2, $3)
          ON CONFLICT("id") DO UPDATE SET first_name=$2, last_name=$3 RETURNING id`

	_, databaseErr := individualRepo.databaseConnection.Exec(
		context.Background(),
		query,
		id,
		firstName,
		lastName,
	)

	return databaseErr
}

func (individualRepo *individualRepoImp) UpdateOrInsert(id int64, firstName string, lastName string) error {
	query := `UPDATE individuals SET first_name=$2, last_name=$3 WHERE id = $1`

	t, databaseErr := individualRepo.databaseConnection.Exec(
		context.Background(),
		query,
		id,
		firstName,
		lastName,
	)

	if t.RowsAffected() == 0 {
		return individualRepo.Insert(id, firstName, lastName)
	}

	return databaseErr
}

func (individualRepo *individualRepoImp) BatchInsert(batchData []map[string]string) {
	query := `INSERT INTO individuals(id, first_name, last_name)
          VALUES($1, $2, $3)
          ON CONFLICT("id") DO UPDATE SET first_name=$2, last_name=$3 RETURNING id`

	batch := pgx.Batch{}
	for _, v := range batchData {
		batch.Queue(query, v["id"], v["first_name"], v["last_name"])
	}

	bc := individualRepo.databaseConnection.SendBatch(
		context.Background(),
		&batch,
	)

	bc.Close()
}

func (individualRepo *individualRepoImp) GetNamesBySingleString(name string, searchType domain.SearchType) ([]domain.Individual, error) {
	var result []domain.Individual
	var queryStringBuilder strings.Builder
	queryStringBuilder.WriteString("SELECT id, first_name, last_name from individuals WHERE ")

	if searchType == domain.Weak || searchType == domain.Both {
		likeString := "CONCAT('%',LOWER(@name),'%')"

		queryStringBuilder.WriteString(
			fmt.Sprintf(
				" LOWER(first_name) LIKE %s OR LOWER(last_name) LIKE %s ",
				likeString,
				likeString,
			),
		)
	}

	if searchType == domain.Strong || searchType == domain.Both {
		if searchType == domain.Both {
			queryStringBuilder.WriteString(" OR ")
		}

		queryStringBuilder.WriteString(
			"LOWER(first_name) = LOWER(@name) OR LOWER(last_name) = LOWER(@name)",
		)
	}

	rows, err := individualRepo.databaseConnection.Query(context.Background(), queryStringBuilder.String(), pgx.NamedArgs{
		"name": name,
	})

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		individual := domain.Individual{}
		err = rows.Scan(&individual.Uid, &individual.FirstName, &individual.LastName)
		if err != nil {
			return nil, err
		}

		result = append(result, individual)
	}

	return result, nil
}

func (individualRepo *individualRepoImp) GetNamesByFirstAndLastName(firstName, lastName string, searchType domain.SearchType) ([]domain.Individual, error) {
	var result []domain.Individual
	var queryStringBuilder strings.Builder
	queryStringBuilder.WriteString("SELECT id, first_name, last_name from individuals WHERE ")

	if searchType == domain.Weak || searchType == domain.Both {
		firstNameLikeString := "CONCAT('%',LOWER(@first_name),'%')"
		lastNamelikeString := "CONCAT('%',LOWER(@last_name),'%')"

		queryStringBuilder.WriteString(
			fmt.Sprintf(
				`(LOWER(first_name) LIKE %s OR LOWER(last_name) LIKE %s) OR 
				(LOWER(first_name) LIKE %s AND LOWER(last_name) LIKE %s)  
				`,
				firstNameLikeString,
				lastNamelikeString,
				lastNamelikeString,
				firstNameLikeString,
			),
		)
	}

	if searchType == domain.Strong || searchType == domain.Both {
		if searchType == domain.Both {
			queryStringBuilder.WriteString(" OR ")
		}

		queryStringBuilder.WriteString(
			fmt.Sprintf(
				`(LOWER(first_name) = LOWER(@first_name) AND LOWER(last_name) = LOWER(@last_name)) 
					OR (LOWER(first_name) = LOWER(@last_name) AND LOWER(last_name) = LOWER(@first_name))`,
			),
		)
	}

	rows, err := individualRepo.databaseConnection.Query(context.Background(), queryStringBuilder.String(), pgx.NamedArgs{
		"first_name": firstName,
		"last_name":  lastName,
	})

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		individual := domain.Individual{}
		err = rows.Scan(&individual.Uid, &individual.FirstName, &individual.LastName)
		if err != nil {
			return nil, err
		}

		result = append(result, individual)
	}

	return result, nil
}
