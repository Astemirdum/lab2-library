package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Astemirdum/library-service/identity-provider/internal/model"
	"go.uber.org/zap"
)

type repository struct {
	db  *pgxpool.Pool
	log *zap.Logger
}

func NewRepository(db *pgxpool.Pool, log *zap.Logger) (*repository, error) {
	return &repository{
		db:  db,
		log: log.Named("repo"),
	}, nil
}

type Repository interface {
	Create(ctx context.Context, user model.User) error
	Get(ctx context.Context, username string) (model.User, error)
}

func (r *repository) Create(ctx context.Context, user model.User) error {
	q := `insert into users (username, password, email) 
	values (@username, @password, @email) on conflict do nothing`
	args := pgx.NamedArgs{
		"username": user.Username,
		"password": user.Password,
		"email":    user.Email,
	}
	if _, err := r.db.Exec(ctx, q, args); err != nil {
		return err
	}
	return nil
}

func (r *repository) Get(ctx context.Context, username string) (model.User, error) {
	const q = `
	select  id, username, password, email, user_type  from users
	where username = @username`
	args := pgx.NamedArgs{"username": username}
	rows, err := r.db.Query(ctx, q, args)
	if err != nil {
		return model.User{}, err
	}
	defer rows.Close()
	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[model.User])
	if err != nil {
		return model.User{}, fmt.Errorf("pgx.CollectRows: %w", err)
	}
	return user, nil
}
