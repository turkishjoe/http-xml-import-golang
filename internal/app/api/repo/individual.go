package repo

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
)

type IndividualRepo interface {
	IsEmpty() error
	Insert(id int64, firstName string, lastName string) bool
	BatchInsert(batchData []map[string]string)
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
	query := IndividualRepo.buildInsertQuery()

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
	query := IndividualRepo.buildInsertQuery()

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

func (IndividualRepo *individualRepoImp) buildInsertQuery() string {
	sb := strings.Builder{}

	sb.WriteString("INSERT INTO individuals(id, first_name, last_name) ")
	sb.WriteString(" VALUES($1, $2, $3) ")
	sb.WriteString("ON CONFLICT(\"id\") DO UPDATE SET first_name=$2, last_name=$3 RETURNING id")

	return sb.String()
}
