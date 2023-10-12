package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Astemirdum/library-service/reservation/internal/errs"

	"github.com/Astemirdum/library-service/reservation/internal/model"
	"github.com/google/uuid"

	sq "github.com/Masterminds/squirrel"
	"go.uber.org/zap"
)

type Repository interface {
	CreateReservation(ctx context.Context, req model.CreateReservationRequest) (model.Reservation, error)
	GetRented(ctx context.Context, username string) (int, error)
	GetReservations(ctx context.Context, username string) ([]model.Reservation, error)
	ReservationsReturn(ctx context.Context, username, reservationUid string) (model.ReservationReturnResponse, error)
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
	reservationTableName = `reservation`
)

var qb = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

func (r *repository) ReservationsReturn(ctx context.Context, username, reservationUid string) (model.ReservationReturnResponse, error) {
	q := fmt.Sprintf(`update %s
	set status = case when date(now()) > till_date 
	    then 'EXPIRED' else 'RETURNED' end
	where reservation_uid = @reservation_uid and username = @username and status='RENTED'
	returning library_uid, book_uid`, reservationTableName)

	args := pgx.NamedArgs{
		"reservation_uid": reservationUid,
		"username":        username,
	}
	var resp model.ReservationReturnResponse
	err := r.db.QueryRow(ctx, q, args).Scan(&resp.LibraryUid, &resp.BookUid)
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
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[model.Reservation])
}

func (r *repository) GetRented(ctx context.Context, username string) (int, error) {
	q := `
	select count(*) from reservation
	where username = $1 and status = 'RENTED'
`
	var count int
	if err := r.db.QueryRow(ctx, q, username).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *repository) CreateReservation(ctx context.Context, req model.CreateReservationRequest) (model.Reservation, error) {
	q := fmt.Sprintf(`insert into %s (reservation_uid, username, book_uid, library_uid, status, start_date, till_date) 
	values (@reservation_uid, @username, @book_uid, @library_uid, @status, now(), @till_date) 
	returning *`, reservationTableName)
	batch := &pgx.Batch{}
	ids := []uuid.UUID{uuid.New()}
	for _, id := range ids {
		args := pgx.NamedArgs{
			"reservation_uid": id,
			"username":        req.UserName,
			"book_uid":        req.BookUid,
			"library_uid":     req.LibraryUid,
			"status":          model.StatusRented,
			"till_date":       req.TillDate.Format(time.DateOnly),
		}
		batch.Queue(q, args)
	}

	result := r.db.SendBatch(ctx, batch)
	defer result.Close()

	ress := make([]model.Reservation, 0, len(ids))
	for _, uid := range ids {
		rows, err := result.Query()
		var pgErr *pgconn.PgError
		if err != nil {
			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
				r.log.Warn("reservation_uid already exists", zap.Any("reservation_uid", uid))
			} else {
				return model.Reservation{}, fmt.Errorf("unable to insert row: %w", err)
			}
		}
		res, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[model.Reservation])
		if err != nil {
			return model.Reservation{}, fmt.Errorf("pgx.CollectRows: %w", err)
		}
		ress = append(ress, res)
		rows.Close()
	}

	return ress[0], result.Close()
}

// CreateReservationCopy  insert bacth without check.
func (r *repository) CreateReservationCopy(ctx context.Context, req model.CreateReservationRequest) error {
	entries := [][]any{}
	columns := []string{"reservation_uid", "username", "book_uid", "library_uid", "status", "start_date", "till_date"}
	ids := []uuid.UUID{uuid.New()}
	for _, id := range ids {
		entries = append(entries, []any{id, req.UserName, req.BookUid, req.LibraryUid, model.StatusRented, time.Now().UTC(), req.TillDate.Format(time.DateOnly)})
	}

	if _, err := r.db.CopyFrom(
		ctx,
		pgx.Identifier{reservationTableName},
		columns,
		pgx.CopyFromRows(entries),
	); err != nil {
		return fmt.Errorf("error copying into %s table: %w", reservationTableName, err)
	}

	return nil
}
