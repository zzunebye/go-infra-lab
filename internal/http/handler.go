package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"go-infra-lab/internal/link"
)

const maxRequestBytes = 1 << 20

type LinkService interface {
	CreateLink(ctx context.Context, originalURL string) (link.Link, error)
	GetLink(ctx context.Context, code string) (link.Link, error)
	ResolveLink(ctx context.Context, code string) (string, error)
}

type Handler struct {
	service LinkService
	baseURL string
}

func NewHandler(service LinkService, baseURL string) *Handler {
	return &Handler{service: service, baseURL: strings.TrimRight(baseURL, "/")}
}

func (h *Handler) CreateLink(w http.ResponseWriter, r *http.Request) {
	if !requireJSON(w, r) {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request struct {
		URL string `json:"url"`
	}
	if err := decoder.Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request")
		return
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "request must contain one JSON object")
		return
	}

	item, err := h.service.CreateLink(r.Context(), request.URL)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, struct {
		Code        string `json:"code"`
		ShortURL    string `json:"shortUrl"`
		OriginalURL string `json:"originalUrl"`
	}{Code: item.Code, ShortURL: h.baseURL + "/" + item.Code, OriginalURL: item.OriginalURL})
}

func (h *Handler) GetLink(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.GetLink(r.Context(), r.PathValue("code"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	target, err := h.service.ResolveLink(r.Context(), r.PathValue("code"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	http.Redirect(w, r, target, http.StatusFound)
}

func requireJSON(w http.ResponseWriter, r *http.Request) bool {
	mediaType := strings.ToLower(strings.TrimSpace(strings.Split(r.Header.Get("Content-Type"), ";")[0]))
	if mediaType != "application/json" {
		writeError(w, http.StatusUnsupportedMediaType, "content type must be application/json")
		return false
	}
	return true
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, link.ErrInvalidURL):
		writeError(w, http.StatusBadRequest, link.ErrInvalidURL.Error())
	case errors.Is(err, link.ErrLinkNotFound):
		writeError(w, http.StatusNotFound, link.ErrLinkNotFound.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, struct {
		Error string `json:"error"`
	}{Error: message})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
