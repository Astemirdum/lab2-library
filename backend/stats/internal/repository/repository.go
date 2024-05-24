package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	statsModel "github.com/Astemirdum/library-service/backend/pkg/kafka"
	"github.com/Astemirdum/library-service/backend/stats/internal/model"
	"go.uber.org/zap"
)

type Repository interface {
	GetStats(ctx context.Context) (model.StatsInfo, error)
	Stats(ctx context.Context, events statsModel.EventStats) error
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

func (r *repository) Stats(ctx context.Context, event statsModel.EventStats) error {
	q := `insert into events (timestamp, username, reservation_uid, book_uid, library_uid, event_type, simplex, rating) 
	values (@timestamp, @username, @reservation_uid, @book_uid, @library_uid, @event_type, @simplex, @rating)`
	args := pgx.NamedArgs{
		"timestamp":       event.Timestamp,
		"username":        event.UserName,
		"reservation_uid": event.ReservationID,
		"book_uid":        event.BookID,
		"library_uid":     event.LibraryID,
		"rating":          event.Rating,
		"event_type":      event.EventType,
		"simplex":         event.Simplex,
	}
	_, err := r.db.Exec(ctx, q, args)
	return err
}

func (r *repository) GetStats(ctx context.Context) (model.StatsInfo, error) {
	const q = `
	select username, max(timestamp) as last_updated, (avg(rating) filter(where rating > 0))::int as rating,  
	       coalesce(count(distinct reservation_uid) filter ( where  simplex = 'UP'), 0) - coalesce(count(distinct reservation_uid) filter ( where  simplex = 'DOWN'), 0) as cnt_reserv, 
	       coalesce(count(book_uid) filter ( where  simplex = 'UP'), 0) - coalesce(count(book_uid) filter ( where  simplex = 'DOWN'), 0) as cnt_books,  
	       coalesce(count(library_uid) filter ( where  simplex = 'UP'), 0) - coalesce(count(library_uid) filter ( where  simplex = 'DOWN'), 0) as cnt_libs 
	from events
	group by username
	order by username
`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return model.StatsInfo{}, err
	}
	defer rows.Close()
	stats, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.Stats])
	if err != nil {
		return model.StatsInfo{}, fmt.Errorf("pgx.CollectRows: %w", err)
	}
	return model.StatsInfo{Data: stats}, nil
}
