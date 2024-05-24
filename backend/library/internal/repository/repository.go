package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Astemirdum/library-service/backend/library/internal/errs"
	"github.com/pkg/errors"

	sq "github.com/Masterminds/squirrel"

	"github.com/Astemirdum/library-service/backend/library/internal/model"
	"go.uber.org/zap"
)

type Repository interface {
	ListLibrary(ctx context.Context, city string, page, size int) (model.ListLibraries, error)
	ListBooks(ctx context.Context, libraryUid string, showAll bool, page, size int) (model.ListBooks, error)
	GetBook(ctx context.Context, libraryUid, bookUid string) (model.Book, error)
	GetLibrary(ctx context.Context, libraryUid string) (model.Library, error)
	AvailableCount(ctx context.Context, libraryID, bookID int, isReturn bool) error
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
	libraryTableName      = `library`
	booksTableName        = `books`
	libraryBooksTableName = `library_books`
)

var qb = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

func (r *repository) GetBook(ctx context.Context, libraryUid, bookUid string) (model.Book, error) {
	query, args, err := qb.Select("b.id", "book_uid", "b.name", "author", "genre", "condition", "available_count").
		From(booksTableName + " b").
		Join(fmt.Sprintf("%s lb on b.id = lb.book_id", libraryBooksTableName)).
		Join(fmt.Sprintf("%s l on l.id = lb.library_id", libraryTableName)).
		Where(sq.Eq{"library_uid": libraryUid}).
		Where(sq.Eq{"book_uid": bookUid}).
		Limit(1).
		ToSql()
	if err != nil {
		return model.Book{}, err
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return model.Book{}, err
	}
	defer rows.Close()

	book, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[model.Book])
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Book{}, errs.ErrNotFound
		}
		return model.Book{}, err
	}

	return book, nil
}

func (r *repository) AvailableCount(ctx context.Context, libraryID, bookID int, isReturn bool) error {
	q := `
update library_books
    set available_count = available_count + @inc
where library_id = @library_id and book_id = @book_id`
	inc := 1
	if !isReturn {
		inc = -1
	}
	args := pgx.NamedArgs{
		"library_id": libraryID,
		"book_id":    bookID,
		"inc":        inc,
	}
	_, err := r.db.Exec(ctx, q, args)
	return err
}

func (r *repository) GetLibrary(ctx context.Context, libraryUid string) (model.Library, error) {
	query, args, err := qb.Select("id", "library_uid", "name", "city", "address").
		From(libraryTableName).
		Where(sq.Eq{"library_uid": libraryUid}).
		Limit(1).
		ToSql()
	if err != nil {
		return model.Library{}, err
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return model.Library{}, err
	}
	defer rows.Close()

	lib, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[model.Library])
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Library{}, errs.ErrNotFound
		}
		return model.Library{}, err
	}

	return lib, nil
}

func (r *repository) ListLibrary(ctx context.Context, city string, page, size int) (model.ListLibraries, error) {
	q := qb.Select("id", "library_uid", "name", "city", "address").
		From(libraryTableName).
		Where(sq.Eq{"city": city})

	if page != 0 && size != 0 {
		q = q.Limit(uint64(size)).Offset(uint64((page - 1) * size))
	}

	query, args, err := q.ToSql()
	if err != nil {
		return model.ListLibraries{}, err
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return model.ListLibraries{}, err
	}
	defer rows.Close()

	libs, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.Library])
	if err != nil {
		return model.ListLibraries{}, fmt.Errorf("pgx.CollectRows: %w", err)
	}
	return model.ListLibraries{
		Paging: model.Paging{
			Page:          page,
			PageSize:      size,
			TotalElements: len(libs),
		},
		Items: libs,
	}, nil
}

func (r *repository) ListBooks(ctx context.Context, libraryUid string, showAll bool, page, size int) (model.ListBooks, error) {
	q := qb.Select("b.id", "book_uid", "b.name", "author", "genre", "condition", "available_count").
		From(booksTableName + " b").
		Join(fmt.Sprintf("%s lb on b.id = lb.book_id", libraryBooksTableName)).
		Join(fmt.Sprintf("%s l on l.id = lb.library_id", libraryTableName)).
		Where(sq.Eq{"library_uid": libraryUid})

	if !showAll {
		q = q.Where(sq.Gt{"available_count": 0})
	}
	if page != 0 && size != 0 {
		q = q.Limit(uint64(size)).Offset(uint64((page - 1) * size))
	}

	query, args, err := q.ToSql()
	if err != nil {
		return model.ListBooks{}, err
	}
	r.log.Debug("ListBooks", zap.String("query", query), zap.Any("args", args))

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return model.ListBooks{}, err
	}
	defer rows.Close()

	books, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.Book])
	if err != nil {
		return model.ListBooks{}, err
	}

	return model.ListBooks{
		Paging: model.Paging{
			Page:          page,
			PageSize:      size,
			TotalElements: len(books),
		},
		Items: books,
	}, nil
}
