package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go-infra-lab/internal/link"
	"go-infra-lab/internal/storage"
)

func TestLinkEndpoints(t *testing.T) {
	service := link.NewService(storage.NewMemoryRepository())
	handler := NewHandler(service, "http://localhost:8080")
	server := httptest.NewServer(NewRouter(handler))
	defer server.Close()

	request, err := http.NewRequest(http.MethodPost, server.URL+"/links", strings.NewReader(`{"url":"https://example.com/path"}`))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("POST /links status = %d", response.StatusCode)
	}
	var created struct {
		Code        string `json:"code"`
		ShortURL    string `json:"shortUrl"`
		OriginalURL string `json:"originalUrl"`
	}
	if err := json.NewDecoder(response.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}
	if created.Code == "" || created.ShortURL != "http://localhost:8080/"+created.Code || created.OriginalURL != "https://example.com/path" {
		t.Fatalf("unexpected creation response: %+v", created)
	}

	for _, test := range []struct {
		path string
		want int
	}{{"/links/" + created.Code, http.StatusOK}, {"/" + created.Code, http.StatusFound}, {"/links/missing", http.StatusNotFound}} {
		client := server.Client()
		client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }
		response, err := client.Get(server.URL + test.path)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != test.want {
			t.Errorf("GET %s status = %d, want %d", test.path, response.StatusCode, test.want)
		}
		if test.want == http.StatusFound && response.Header.Get("Location") != "https://example.com/path" {
			t.Errorf("redirect Location = %q", response.Header.Get("Location"))
		}
	}
}

func TestCreateLinkRejectsInvalidURL(t *testing.T) {
	handler := NewHandler(link.NewService(storage.NewMemoryRepository()), "http://localhost:8080")
	request := httptest.NewRequest(http.MethodPost, "/links", strings.NewReader(`{"url":"ftp://example.com"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	NewRouter(handler).ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}
