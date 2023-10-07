package repository

import (
	"context"
	"github.com/Astemirdum/library-service/reservation/internal/model"
	"github.com/google/uuid"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type Repository interface {
	CreateReservation(ctx context.Context, req model.CreateReservationRequest) (model.Reservation, error)
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
	reservationTableName = `reservation`
)

var qb = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

func (r *repository) CreateReservation(ctx context.Context, req model.CreateReservationRequest) (model.Reservation, error) {
	q, args, err := qb.Insert(reservationTableName).
		Columns("reservation_uid", "username", "book_uid", "library_uid", "status", "start_date", "till_date").
		Values(uuid.New(), req.UserName, req.BookUid, req.LibraryUid, model.StatusRented, time.Now().UTC(), req.TillDate.Format(time.DateOnly)).
		Suffix("returning *").
		ToSql()
	if err != nil {
		return model.Reservation{}, err
	}
	var res model.Reservation
	if err := r.db.GetContext(ctx, &res, q, args...); err != nil {
		r.log.Error("CreateReservation", zap.String("q", q), zap.Any("args", args))
		return model.Reservation{}, err
	}
	return res, nil
}
