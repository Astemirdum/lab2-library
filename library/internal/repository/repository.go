package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Astemirdum/library-service/library/internal/errs"
	"github.com/pkg/errors"

	sq "github.com/Masterminds/squirrel"

	"github.com/Astemirdum/library-service/library/internal/model"
	"github.com/jmoiron/sqlx"
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
	libraryTableName      = `library`
	booksTableName        = `books`
	libraryBooksTableName = `library_books`
)

var qb = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

func (r *repository) GetBook(ctx context.Context, libraryUid, bookUid string) (model.Book, error) {
	query, args, err := qb.Select("b.id", "book_uid", "b.name", "author", "genre", "condition").
		From(booksTableName + " b").
		Join(fmt.Sprintf("%s lb on b.id = lb.book_id", libraryBooksTableName)).
		Join(fmt.Sprintf("%s l on l.id = lb.library_id", libraryTableName)).
		Where(sq.Eq{"library_uid": libraryUid}).
		Where(sq.Eq{"book_uid": bookUid}).
		Where(sq.Gt{"available_count": 0}).
		Limit(1).
		ToSql()
	if err != nil {
		return model.Book{}, err
	}

	var book model.Book
	if err := r.db.GetContext(ctx, &book, query, args...); err != nil {
		r.log.Error("GetBook", zap.String("q", query), zap.Any("args", args))
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
    set available_count = available_count + $3
where library_id = $1 and book_id = $2`
	inc := 1
	if !isReturn {
		inc = -1
	}
	_, err := r.db.ExecContext(ctx, q, libraryID, bookID, inc)
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

	var lib model.Library
	if err := r.db.GetContext(ctx, &lib, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Library{}, errs.ErrNotFound
		}
		return model.Library{}, err
	}

	return lib, nil
}

func (r *repository) ListLibrary(ctx context.Context, city string, page, size int) (model.ListLibraries, error) {
	q := qb.Select("library_uid", "name", "city", "address").
		From(libraryTableName).
		Where(sq.Eq{"city": city})

	if page != 0 && size != 0 {
		q = q.Limit(uint64(size)).Offset(uint64((page - 1) * size))
	}

	query, args, err := q.ToSql()
	if err != nil {
		return model.ListLibraries{}, err
	}

	var libs []model.Library
	if err := r.db.SelectContext(ctx, &libs, query, args...); err != nil {
		return model.ListLibraries{}, err
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
	q := qb.Select("book_uid", "b.name", "author", "genre", "condition", "available_count").
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

	var books []model.Book
	if err := r.db.SelectContext(ctx, &books, query, args...); err != nil {
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
