package pagination

import "testing"

func TestNormalize(t *testing.T) {
	p := Normalize(Page{Page: 0, PageSize: 0})
	if p.Page != 1 || p.PageSize != defaultPageSize || p.Offset != 0 {
		t.Fatalf("got %+v", p)
	}
	p = Normalize(Page{Page: 2, PageSize: 10})
	if p.Offset != 10 {
		t.Fatalf("offset got %d", p.Offset)
	}
}

func TestTotalPages(t *testing.T) {
	if TotalPages(125, 20) != 7 {
		t.Fatal("expected 7 pages")
	}
	if TotalPages(0, 20) != 0 {
		t.Fatal("expected 0")
	}
}

func TestMetaFrom(t *testing.T) {
	m := MetaFrom(1, 20, 125)
	if m.TotalPages != 7 || m.Total != 125 {
		t.Fatalf("got %+v", m)
	}
}
