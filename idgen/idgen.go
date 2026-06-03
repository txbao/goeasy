package idgen

import (
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	"github.com/txbao/goeasy/config"
)

var snowflakeSeq uint64

// Generator ID 生成器。
type Generator struct {
	mode string
}

func New(cfg config.IDGenCfg) *Generator {
	mode := cfg.Mode
	if mode == "" {
		mode = "uuid"
	}
	return &Generator{mode: mode}
}

func (g *Generator) NextID() string {
	switch g.mode {
	case "snowflake":
		return nextSnowflake()
	default:
		return uuid.NewString()
	}
}

func nextSnowflake() string {
	seq := atomic.AddUint64(&snowflakeSeq, 1)
	ms := uint64(time.Now().UnixMilli())
	return formatSnowflake(ms, seq)
}

func formatSnowflake(ms, seq uint64) string {
	buf := make([]byte, 16)
	for i := range 8 {
		buf[i] = byte(ms >> (56 - i*8))
	}
	for i := range 8 {
		buf[8+i] = byte(seq >> (56 - i*8))
	}
	return uuid.NewSHA1(uuid.NameSpaceOID, buf).String()
}
