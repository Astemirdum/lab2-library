package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Astemirdum/library-service/backend/rating/internal/errs"
	"github.com/pkg/errors"

	sq "github.com/Masterminds/squirrel"

	"github.com/Astemirdum/library-service/backend/rating/internal/model"
	"go.uber.org/zap"
)

type Repository interface {
	GetRating(ctx context.Context, name string) (model.Rating, error)
	Rating(ctx context.Context, name string, stars int) error
	CreateRating(ctx context.Context, name string, stars int) error
}

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

const (
	ratingTableName = `rating`
)

var qb = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

func (r *repository) Rating(ctx context.Context, name string, stars int) error {
	q := `
update rating 
set stars = stars + @stars
where username=@username`
	args := pgx.NamedArgs{
		"username": name,
		"stars":    stars,
	}
	_, err := r.db.Exec(ctx, q, args)
	return err
}

func (r *repository) CreateRating(ctx context.Context, name string, stars int) error {
	q := `insert into rating (username, stars) values (@username, @stars) on conflict do nothing`
	args := pgx.NamedArgs{
		"username": name,
		"stars":    stars,
	}
	_, err := r.db.Exec(ctx, q, args)
	return err
}

func (r *repository) GetRating(ctx context.Context, name string) (model.Rating, error) {
	q, args, err := qb.Select("stars").
		From(ratingTableName).
		Where(sq.Eq{"username": name}).
		ToSql()
	if err != nil {
		return model.Rating{}, err
	}

	//var rr model.Rating
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return model.Rating{}, err
	}
	defer rows.Close()
	rr, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[model.Rating])
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Rating{}, errs.ErrNotFound
		}
		return model.Rating{}, fmt.Errorf("pgx.CollectRows: %w", err)
	}
	return rr, nil
}
