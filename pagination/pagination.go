package pagination

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// Page 分页参数。
type Page struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Offset   int `json:"-"`
}

const defaultPageSize = 20
const maxPageSize = 200

// Parse 从 Gin 查询参数解析分页。
func Parse(c *gin.Context) Page {
	p, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	ps, _ := strconv.Atoi(c.DefaultQuery("page_size", strconv.Itoa(defaultPageSize)))
	if p < 1 {
		p = 1
	}
	if ps < 1 {
		ps = defaultPageSize
	}
	if ps > maxPageSize {
		ps = maxPageSize
	}
	return Page{
		Page:     p,
		PageSize: ps,
		Offset:   (p - 1) * ps,
	}
}

// Result 分页结果元数据。
type Result struct {
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
}
