package outbox

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/txbao/goeasy/config"
	"github.com/txbao/goeasy/database"
	"github.com/txbao/goeasy/mq"
)

const (
	statusPending = "pending"
	statusSent    = "sent"
	statusFailed  = "failed"
)

// Publisher 统一 MQ 发布入口；outbox 关闭时直发，开启时在事务内写入 outbox 表。
type Publisher struct {
	mq     mq.MQ
	db     *sqlx.DB
	driver string
	table  string
	cfg    config.OutboxCfg
}

// NewPublisher 创建发布器。
func NewPublisher(cfg *config.Config, m mq.MQ, db database.DB) *Publisher {
	if cfg == nil || m == nil {
		return &Publisher{}
	}
	p := &Publisher{
		mq:     m,
		driver: cfg.DB.Driver,
		table:  cfg.OutboxTable(),
		cfg:    cfg.MQ.Outbox,
	}
	if db != nil {
		p.db = db.SQLX()
	}
	return p
}

// Enabled 是否启用 Outbox。
func (p *Publisher) Enabled() bool {
	return p != nil && p.cfg.Enabled && p.db != nil
}

// Publish 发布消息；outbox 开启时必须在 Transaction 内调用 PublishInTx。
func (p *Publisher) Publish(ctx context.Context, topic string, body []byte) error {
	if p == nil || p.mq == nil {
		return errors.New("outbox: mq not available")
	}
	if p.Enabled() {
		return errors.New("outbox enabled: use PublishInTx inside database.Transaction")
	}
	return p.mq.Publish(ctx, topic, body)
}

// PublishInTx 在事务内写入 outbox 或直发 MQ（outbox 关闭时等同 Publish）。
func (p *Publisher) PublishInTx(ctx context.Context, topic string, body []byte) error {
	if p == nil {
		return errors.New("outbox: publisher nil")
	}
	if topic == "" {
		return errors.New("outbox: topic is empty")
	}
	if !p.Enabled() {
		return p.Publish(ctx, topic, body)
	}
	if !database.InTransaction(ctx) {
		return errors.New("outbox: PublishInTx requires database.Transaction context")
	}
	eventID := uuid.NewString()
	ext := database.ExtContext(ctx, p.db)
	query, args := insertSQL(p.driver, p.table, eventID, topic, body)
	_, err := ext.ExecContext(ctx, query, args...)
	return err
}

func insertSQL(driver, table, eventID, topic string, body []byte) (string, []any) {
	if driver == "mysql" {
		return fmt.Sprintf(
			"INSERT INTO %s (event_id, topic, payload, status, retry_count) VALUES (?, ?, ?, ?, 0)",
			table,
		), []any{eventID, topic, body, statusPending}
	}
	return fmt.Sprintf(
		"INSERT INTO %s (event_id, topic, payload, status, retry_count) VALUES ($1, $2, $3, $4, 0)",
		table,
	), []any{eventID, topic, body, statusPending}
}

// Relay 轮询 outbox 表并投递到 MQ。
type Relay struct {
	pub    *Publisher
	mq     mq.MQ
	db     *sqlx.DB
	driver string
	table  string
	cfg    config.OutboxCfg
}

// NewRelay 创建 relay（需 outbox.enabled）。
func NewRelay(cfg *config.Config, m mq.MQ, db database.DB) *Relay {
	if cfg == nil || !cfg.MQ.Outbox.Enabled {
		return nil
	}
	sqlxDB := db.SQLX()
	if sqlxDB == nil {
		return nil
	}
	return &Relay{
		pub:    NewPublisher(cfg, m, db),
		mq:     m,
		db:     sqlxDB,
		driver: cfg.DB.Driver,
		table:  cfg.OutboxTable(),
		cfg:    cfg.MQ.Outbox,
	}
}

type row struct {
	ID         int64  `db:"id"`
	EventID    string `db:"event_id"`
	Topic      string `db:"topic"`
	Payload    []byte `db:"payload"`
	RetryCount int    `db:"retry_count"`
}

// RunOnce 拉取一批 pending 记录投递 MQ。
func (r *Relay) RunOnce(ctx context.Context) error {
	if r == nil || r.db == nil || r.mq == nil {
		return nil
	}
	rows, err := r.fetchPending(ctx)
	if err != nil {
		return err
	}
	for _, item := range rows {
		if err := r.mq.Publish(ctx, item.Topic, item.Payload); err != nil {
			_ = r.markFailed(ctx, item.ID, item.RetryCount+1)
			continue
		}
		_ = r.markSent(ctx, item.ID)
	}
	return nil
}

func (r *Relay) fetchPending(ctx context.Context) ([]row, error) {
	limit := r.cfg.BatchSize
	if limit <= 0 {
		limit = 50
	}
	var rows []row
	var query string
	if r.driver == "mysql" {
		query = fmt.Sprintf(
			"SELECT id, event_id, topic, payload, retry_count FROM %s WHERE status = ? AND retry_count < ? ORDER BY id LIMIT ?",
			r.table,
		)
		err := r.db.SelectContext(ctx, &rows, query, statusPending, r.cfg.MaxRetries, limit)
		return rows, err
	}
	query = fmt.Sprintf(
		"SELECT id, event_id, topic, payload, retry_count FROM %s WHERE status = $1 AND retry_count < $2 ORDER BY id LIMIT $3",
		r.table,
	)
	err := r.db.SelectContext(ctx, &rows, query, statusPending, r.cfg.MaxRetries, limit)
	return rows, err
}

func (r *Relay) markSent(ctx context.Context, id int64) error {
	if r.driver == "mysql" {
		q := fmt.Sprintf("UPDATE %s SET status = ?, sent_at = ? WHERE id = ?", r.table)
		_, err := r.db.ExecContext(ctx, q, statusSent, time.Now().UTC(), id)
		return err
	}
	q := fmt.Sprintf("UPDATE %s SET status = $1, sent_at = $2 WHERE id = $3", r.table)
	_, err := r.db.ExecContext(ctx, q, statusSent, time.Now().UTC(), id)
	return err
}

func (r *Relay) markFailed(ctx context.Context, id int64, retryCount int) error {
	status := statusPending
	if retryCount >= r.cfg.MaxRetries {
		status = statusFailed
	}
	if r.driver == "mysql" {
		q := fmt.Sprintf("UPDATE %s SET retry_count = ?, status = ? WHERE id = ?", r.table)
		_, err := r.db.ExecContext(ctx, q, retryCount, status, id)
		return err
	}
	q := fmt.Sprintf("UPDATE %s SET retry_count = $1, status = $2 WHERE id = $3", r.table)
	_, err := r.db.ExecContext(ctx, q, retryCount, status, id)
	return err
}

// PollCronSpec 根据轮询间隔生成 cron 表达式（秒级）。
func PollCronSpec(interval time.Duration) string {
	sec := int(interval.Seconds())
	if sec < 1 {
		sec = 5
	}
	return fmt.Sprintf("*/%d * * * * *", sec)
}
