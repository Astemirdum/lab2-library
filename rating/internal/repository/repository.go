package repository

import (
	"context"
	"database/sql"

	"github.com/Astemirdum/library-service/rating/internal/errs"
	"github.com/pkg/errors"

	sq "github.com/Masterminds/squirrel"

	ratingModel "github.com/Astemirdum/library-service/rating/internal/model"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type Repository interface {
	GetRating(ctx context.Context, name string) (ratingModel.Rating, error)
	Rating(ctx context.Context, name string, stars int) error
}

type repository struct {
	db  *sqlx.DB
	log *zap.Logger
}

func NewRepository(db *sqlx.DB, log *zap.Logger) (*repository, error) {
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
set stars = stars + $1
where username=$2`
	_, err := r.db.ExecContext(ctx, q, stars, name)
	return err
}

func (r *repository) GetRating(ctx context.Context, name string) (ratingModel.Rating, error) {
	q, args, err := qb.Select("stars").
		From(ratingTableName).
		Where(sq.Eq{"username": name}).
		ToSql()
	if err != nil {
		return ratingModel.Rating{}, err
	}

	var rr ratingModel.Rating
	if err := r.db.GetContext(ctx, &rr, q, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ratingModel.Rating{}, errs.ErrNotFound
		}
		return ratingModel.Rating{}, err
	}

	return rr, nil
}
