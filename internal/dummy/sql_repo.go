package dummy

import (
	"context"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/jackc/pgx/v5/pgxpool"

	"ws-dummy-go/internal/dummy/domain"
)

var db = goqu.Dialect("postgres")

type UsersSQLRepo interface {
	Insert(string) (domain.UserID, error)
}

func NewUsersSQLRepo(p *pgxpool.Pool) UsersSQLRepo {
	return sqlRepo{
		pool: p,
	}
}

type sqlRepo struct {
	pool *pgxpool.Pool
}

// CREATE TABLE public.users (
// 	user_id bigint GENERATED ALWAYS AS IDENTITY,
// 	"name" varchar NOT NULL,
// 	created_at timestamp NOT NULL,
// 	CONSTRAINT users_pk PRIMARY KEY (user_id)
// );

func (r sqlRepo) Insert(name string) (domain.UserID, error) {
	q1 := db.Insert("users").
		Cols("name", "created_at").
		Vals(goqu.Vals{name, goqu.L("NOW()")})

	sql1, params, err := q1.ToSQL()
	if err != nil {
		return "", fmt.Errorf("creating query: %w", err)
	}
	_, err = r.pool.Exec(context.TODO(), sql1, params...)
	if err != nil {
		return "", fmt.Errorf("executing query: %w", err)
	}

	q2 := db.Select("user_id").From("users").Order(goqu.I("created_at").Desc()).Limit(1)

	sql2, _, err := q2.ToSQL()
	if err != nil {
		return "", fmt.Errorf("creating query: %w", err)
	}
	var res string
	if err := r.pool.QueryRow(context.TODO(), sql2).Scan(&res); err != nil {
		return "", fmt.Errorf("executing query: %w", err)
	}
	return domain.UserID(res), nil
}
