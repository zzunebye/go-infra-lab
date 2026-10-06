package http

import "net/http"

func NewRouter(handler *Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /links", handler.CreateLink)
	mux.HandleFunc("GET /links/{code}", handler.GetLink)
	mux.HandleFunc("GET /{code}", handler.Redirect)
	return mux
}
