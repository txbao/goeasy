package audit

import (
	"fmt"
	"reflect"
	"strings"
)

// SummaryOptions 变更摘要选项。
type SummaryOptions struct {
	Redact RedactOptions
}

// BuildChangeSummary 从 before/after 提取白名单字段并脱敏。
func BuildChangeSummary(before, after any, fields []string, opts SummaryOptions) (beforeOut, afterOut map[string]any, err error) {
	bm, err := toFieldMap(before, fields)
	if err != nil {
		return nil, nil, err
	}
	am, err := toFieldMap(after, fields)
	if err != nil {
		return nil, nil, err
	}
	return RedactMap(bm, opts.Redact), RedactMap(am, opts.Redact), nil
}

func toFieldMap(v any, fields []string) (map[string]any, error) {
	if v == nil {
		return nil, nil
	}
	switch m := v.(type) {
	case map[string]any:
		return pickMapFields(m, fields), nil
	default:
		return pickStructFields(v, fields)
	}
}

func pickMapFields(m map[string]any, fields []string) map[string]any {
	if m == nil {
		return nil
	}
	out := make(map[string]any, len(fields))
	for _, f := range fields {
		if val, ok := m[f]; ok {
			out[f] = val
		}
	}
	return out
}

func pickStructFields(v any, fields []string) (map[string]any, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil, nil
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("audit: unsupported type %T", v)
	}
	rt := rv.Type()
	fieldSet := make(map[string]struct{}, len(fields))
	for _, f := range fields {
		fieldSet[f] = struct{}{}
	}
	out := make(map[string]any)
	for i := 0; i < rt.NumField(); i++ {
		sf := rt.Field(i)
		if !sf.IsExported() {
			continue
		}
		name := fieldName(sf)
		if _, ok := fieldSet[name]; !ok {
			continue
		}
		out[name] = rv.Field(i).Interface()
	}
	return out, nil
}

func fieldName(sf reflect.StructField) string {
	if tag := sf.Tag.Get("json"); tag != "" && tag != "-" {
		parts := strings.Split(tag, ",")
		if parts[0] != "" {
			return parts[0]
		}
	}
	return sf.Name
}
