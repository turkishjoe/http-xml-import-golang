package repo

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/turkishjoe/xml-parser/internal/app/api/domain"
	"strings"
)

type IndividualRepo interface {
	IsEmpty() bool
	Insert(id int64, firstName string, lastName string) error
	UpdateOrInsert(id int64, firstName string, lastName string) error
	BatchInsert(batchData []map[string]string)
	GetNames(name string, searchType domain.SearchType) ([]domain.Individual, error)
}

type individualRepoImp struct {
	databaseConnection *pgxpool.Pool
}

func CreateRepo(DatabaseConnection *pgxpool.Pool) IndividualRepo {
	return &individualRepoImp{
		databaseConnection: DatabaseConnection,
	}
}

func (IndividualRepo *individualRepoImp) IsEmpty() bool {
	var res int64
	err := IndividualRepo.databaseConnection.QueryRow(context.Background(), "select id from individuals limit 1").Scan(&res)

	if err != nil {
		return false
	}

	return true
}

func (IndividualRepo *individualRepoImp) Insert(id int64, firstName string, lastName string) error {
	query := `INSERT INTO individuals(id, first_name, last_name)
          VALUES($1, $2, $3)
          ON CONFLICT("id") DO UPDATE SET first_name=$2, last_name=$3 RETURNING id`

	_, databaseErr := IndividualRepo.databaseConnection.Exec(
		context.Background(),
		query,
		id,
		firstName,
		lastName,
	)

	return databaseErr
}

func (IndividualRepo *individualRepoImp) UpdateOrInsert(id int64, firstName string, lastName string) error {
	query := `UPDATE individuals SET first_name=$2, last_name=$3 WHERE id = $1`

	_, databaseErr := IndividualRepo.databaseConnection.Exec(
		context.Background(),
		query,
		id,
		firstName,
		lastName,
	)

	return databaseErr
}

func (IndividualRepo *individualRepoImp) BatchInsert(batchData []map[string]string) {
	query := `INSERT INTO individuals(id, first_name, last_name)
          VALUES($1, $2, $3)
          ON CONFLICT("id") DO UPDATE SET first_name=$2, last_name=$3 RETURNING id`

	batch := pgx.Batch{}
	for _, v := range batchData {
		batch.Queue(query, v["id"], v["first_name"], v["last_name"])
	}

	bc := IndividualRepo.databaseConnection.SendBatch(
		context.Background(),
		&batch,
	)

	bc.Close()
}

func (individualRepo *individualRepoImp) GetNames(name string, searchType domain.SearchType) ([]domain.Individual, error) {
	var result []domain.Individual
	var queryStringBuilder strings.Builder
	queryStringBuilder.WriteString("SELECT id, first_name, last_name from individuals WHERE ")

	likeString := "CONCAT('%',LOWER(@name),'%')"

	if searchType == domain.Weak || searchType == domain.Both {
		queryStringBuilder.WriteString(
			fmt.Sprintf(
				"LOWER(first_name) LIKE %s OR LOWER(last_name) LIKE %s",
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
			fmt.Sprintf(
				"LOWER(first_name) = %s OR last_name = %s",
				likeString,
				likeString,
			),
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
