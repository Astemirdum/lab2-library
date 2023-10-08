package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Astemirdum/library-service/reservation/internal/errs"

	"github.com/Astemirdum/library-service/reservation/internal/model"
	"github.com/google/uuid"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type Repository interface {
	CreateReservation(ctx context.Context, req model.CreateReservationRequest) (model.Reservation, error)
	GetRented(ctx context.Context, username string) (int, error)
	GetReservations(ctx context.Context, username string) ([]model.Reservation, error)
	ReservationsReturn(ctx context.Context, username, reservationUid string) (model.ReservationReturnResponse, error)
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

func (r *repository) ReservationsReturn(ctx context.Context, username, reservationUid string) (model.ReservationReturnResponse, error) {
	q := fmt.Sprintf(`update %s
	set status = case when date(now()) > till_date 
	    then 'EXPIRED' else 'RETURNED' end
	where reservation_uid = $1 and username = $2 and status='RENTED'
	returning library_uid, book_uid`, reservationTableName)

	var resp model.ReservationReturnResponse
	err := r.db.QueryRowContext(ctx, q, reservationUid, username).Scan(&resp.LibraryUid, &resp.BookUid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.ReservationReturnResponse{}, errs.ErrNotFound
		}
		return model.ReservationReturnResponse{}, err
	}
	return resp, nil
}

func (r *repository) GetReservations(ctx context.Context, username string) ([]model.Reservation, error) {
	q, args, err := qb.Select("id", "reservation_uid", "username", "book_uid", "library_uid", "status", "start_date", "till_date").
		From(reservationTableName).
		Where(sq.Eq{"username": username}).
		ToSql()
	if err != nil {
		return nil, err
	}
	var items []model.Reservation
	if err := r.db.SelectContext(ctx, &items, q, args...); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *repository) GetRented(ctx context.Context, username string) (int, error) {
	q := `
	select count(*) from reservation
	where username = $1 and status = 'RENTED'
`
	var count int
	if err := r.db.QueryRowContext(ctx, q, username).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

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
