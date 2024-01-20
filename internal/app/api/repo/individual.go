package repo

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IndividualRepo struct {
	DatabaseConnection *pgxpool.Pool
}

func CreateRepo(DatabaseConnection *pgxpool.Pool) IndividualRepo {
	return IndividualRepo{
		DatabaseConnection: DatabaseConnection,
	}
}

func (IndividualRepo *IndividualRepo) Insert() {
	query := "INSERT INTO individuals(id, first_name, last_name) " +
		"VALUES(@id, @first_name, @last_name) " +
		"ON CONFLICT(\"id\") DO UPDATE SET" +
		" first_name=EXCLUDED.first_name, last_name=EXCLUDED.last_name " +
		"RETURNING id"

	_, databaseErr := IndividualRepo.DatabaseConnection.Exec(
		context.Background(),
		query,
		args,
	)

	return databaseErr
}
