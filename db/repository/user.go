package repository

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DbUser struct {
	Id       string
	Username string
	Password string
}

type UserRepository interface {
	GetByUserName(userName string) (DbUser, error)
}

type PgUserRepository struct {
	pool *pgxpool.Pool
}

func NewPgUserRepository(pool *pgxpool.Pool) *PgUserRepository {
	return &PgUserRepository{pool: pool}
}

func (pg *PgUserRepository) GetByUserName(ctx context.Context, userName string) (*DbUser, error) {
	const query = "SELECT id, username, password FROM users u WHERE u.username = $1"
	rows, err := pg.pool.Query(ctx, query, userName)

	if err != nil {
		return nil, err
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByPos[DbUser])
	return &user, err
}
