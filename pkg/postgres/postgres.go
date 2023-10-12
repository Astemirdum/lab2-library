package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pkg/errors"
	goose "github.com/pressly/goose/v3"
)

type DB struct {
	Host     string `yaml:"host" envconfig:"DB_HOST"`
	Port     int    `yaml:"port" envconfig:"DB_PORT"`
	Username string `yaml:"user" envconfig:"DB_USER"`
	Password string `yaml:"password" envconfig:"DB_PASSWORD"`
	NameDB   string `yaml:"dbname" envconfig:"DB_NAME"`
}

var (
	pgOnce     sync.Once
	pgInstance *pgxpool.Pool
)

func NewPostgresDB(ctx context.Context, cfg *DB, fsys fs.FS) (*pgxpool.Pool, error) {
	if err := migrateSchema(cfg, fsys); err != nil {
		return nil, err
	}
	return newDB(ctx, cfg)
}

func newDB(ctx context.Context, cfg *DB) (*pgxpool.Pool, error) {
	var err error
	pgOnce.Do(func() {
		pgInstance, err = pgxpool.New(ctx, newDBURL(cfg))
		if err != nil {
			panic(err)
		}
		if err = pgInstance.Ping(ctx); err != nil {
			panic(err)
		}
	})

	return pgInstance, nil
}

func newDSN(cfg *DB) string {
	return fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.Username, cfg.NameDB, cfg.Password)
}

func newDBURL(cfg *DB) string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.NameDB)
}

func migrateSchema(cfg *DB, fsys fs.FS) error {
	dsn := newDSN(cfg)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		return fmt.Errorf("migrateSchema ping: %w", err)
	}

	goose.SetBaseFS(fsys)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	if err = goose.Up(db, "sql"); err != nil {
		return errors.Wrap(err, "goose run()")
	}
	return nil
}
