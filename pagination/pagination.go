package pagination

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// Page 分页参数。
type Page struct {
	Page     int `json:"page" validate:"min=1"`
	PageSize int `json:"page_size" validate:"min=1,max=200"`
	Offset   int `json:"-"`
}

// Meta 分页结果元数据（用于列表接口 data.pagination）。
type Meta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// Result 分页结果元数据（兼容旧字段，不含 total_pages）。
type Result struct {
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
}

const defaultPageSize = 20
const maxPageSize = 200

// Parse 从 Gin 查询参数解析分页。
func Parse(c *gin.Context) Page {
	p, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	ps, _ := strconv.Atoi(c.DefaultQuery("page_size", strconv.Itoa(defaultPageSize)))
	return Normalize(Page{Page: p, PageSize: ps})
}

// Normalize 校正页码与 page_size，并计算 Offset。
func Normalize(p Page) Page {
	page := p.Page
	ps := p.PageSize
	if page < 1 {
		page = 1
	}
	if ps < 1 {
		ps = defaultPageSize
	}
	if ps > maxPageSize {
		ps = maxPageSize
	}
	return Page{
		Page:     page,
		PageSize: ps,
		Offset:   (page - 1) * ps,
	}
}

// TotalPages 根据总数与每页条数计算总页数。
func TotalPages(total int64, pageSize int) int {
	if pageSize <= 0 || total <= 0 {
		return 0
	}
	return int((total + int64(pageSize) - 1) / int64(pageSize))
}

// MetaFrom 由分页参数与总数构造 Meta。
func MetaFrom(page, pageSize int, total int64) Meta {
	p := Normalize(Page{Page: page, PageSize: pageSize})
	return Meta{
		Page:       p.Page,
		PageSize:   p.PageSize,
		Total:      total,
		TotalPages: TotalPages(total, p.PageSize),
	}
}

// ToResult 将 Meta 转为旧版 Result（不含 total_pages）。
func (m Meta) ToResult() Result {
	return Result{Page: m.Page, PageSize: m.PageSize, Total: m.Total}
}
