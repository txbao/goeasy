package audit

import (
	"context"
	"time"

	"github.com/txbao/goeasy/contextx"
)

// Entry 业务操作日志条目（框架不解释业务字典语义）。
type Entry struct {
	OperatedAt    time.Time
	ModuleCode    string
	ActionType    string
	ObjectType    string
	ObjectID      string
	ObjectName    string
	CustomerID    int64
	SourceAppID   int64
	CurrentOrgID  int64
	Result        string
	FailReason    string
	BeforeSummary map[string]any
	AfterSummary  map[string]any
	Remark        string
	IsSensitive   bool
}

// Recorder 业务操作日志持久化 Port（由业务服务实现 DB 写入）。
type Recorder interface {
	Record(ctx context.Context, op contextx.OperatorContext, entry Entry) error
}

// NopRecorder 默认空实现（零开销）。
type NopRecorder struct{}

func (NopRecorder) Record(context.Context, contextx.OperatorContext, Entry) error {
	return nil
}
