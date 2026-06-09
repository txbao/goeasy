package pagination

// SQLLimitOffset 返回 LIMIT/OFFSET 参数（供 goqu/sqlx 列表查询）。
func SQLLimitOffset(p Page) (limit, offset int) {
	n := Normalize(p)
	return n.PageSize, n.Offset
}
