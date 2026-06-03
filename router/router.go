package router

import "net/http"

func New() *http.ServeMux {
	return http.NewServeMux()
}
