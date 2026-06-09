package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/txbao/goeasy/config"
)

type ctxKey struct{}

// DB 数据库抽象（默认 orm=sqlx）。
type DB interface {
	Ping(ctx context.Context) error
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
	Close() error
	// SQLX 在 enabled 且 orm=sqlx 时返回连接池；Noop 或未启用时为 nil。
	SQLX() *sqlx.DB
}

func Open(cfg config.DB) (DB, error) {
	if !cfg.Enabled {
		return NewNoop(), nil
	}
	if cfg.DSN == "" {
		return nil, errors.New("database enabled but dsn is empty")
	}
	orm := cfg.ORM
	if orm == "" {
		orm = "sqlx"
	}
	if orm != "sqlx" {
		return nil, fmt.Errorf("database.orm %q not implemented (use sqlx)", orm)
	}
	driver := cfg.Driver
	if driver == "" {
		driver = "postgres"
	}
	sqlDriver, err := sqlDriverName(driver)
	if err != nil {
		return nil, err
	}
	db, err := sqlx.Connect(sqlDriver, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("database connect: %w", err)
	}
	if cfg.MaxOpen > 0 {
		db.SetMaxOpenConns(cfg.MaxOpen)
	}
	if cfg.MaxIdle > 0 {
		db.SetMaxIdleConns(cfg.MaxIdle)
	}
	return &sqlxDB{db: db}, nil
}

func sqlDriverName(driver string) (string, error) {
	switch driver {
	case "postgres":
		return "pgx", nil
	case "mysql":
		return "mysql", nil
	default:
		return "", fmt.Errorf("unsupported database.driver %q (use postgres or mysql)", driver)
	}
}

type sqlxDB struct {
	db *sqlx.DB
}

func (d *sqlxDB) SQLX() *sqlx.DB { return d.db }

func (d *sqlxDB) Ping(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

func (d *sqlxDB) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := d.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err := fn(context.WithValue(ctx, ctxKey{}, tx)); err != nil {
		return err
	}
	return tx.Commit()
}

func (d *sqlxDB) Close() error {
	if d.db == nil {
		return nil
	}
	return d.db.Close()
}

// ExtContext 返回事务或连接池，供仓储在 Transaction 内使用。
func ExtContext(ctx context.Context, db *sqlx.DB) sqlx.ExtContext {
	if tx, ok := ctx.Value(ctxKey{}).(*sqlx.Tx); ok {
		return tx
	}
	return db
}

// InTransaction 当前 context 是否处于数据库事务中。
func InTransaction(ctx context.Context) bool {
	_, ok := ctx.Value(ctxKey{}).(*sqlx.Tx)
	return ok
}

type noopDB struct{}

func NewNoop() DB { return &noopDB{} }

func (n *noopDB) SQLX() *sqlx.DB { return nil }

func (n *noopDB) Ping(ctx context.Context) error { return nil }

func (n *noopDB) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func (n *noopDB) Close() error { return nil }

func Validate(cfg config.DB) error {
	if !cfg.Enabled {
		return nil
	}
	if cfg.DSN == "" {
		return fmt.Errorf("database.enabled=true requires database.dsn")
	}
	orm := cfg.ORM
	if orm == "" {
		orm = "sqlx"
	}
	if orm != "sqlx" {
		return fmt.Errorf("database.orm %q not implemented", orm)
	}
	driver := cfg.Driver
	if driver == "" {
		driver = "postgres"
	}
	if _, err := sqlDriverName(driver); err != nil {
		return err
	}
	return nil
}

// RawDB 返回标准库 *sql.DB（可选，用于第三方库）。
func RawDB(d DB) *sql.DB {
	if x, ok := d.(*sqlxDB); ok && x.db != nil {
		return x.db.DB
	}
	return nil
}
