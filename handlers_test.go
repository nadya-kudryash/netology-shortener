package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestShortenHandler(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		wantStatus   int
		wantShortURL bool   // ожидаем непустой short_url в ответе
		wantOriginal string // ожидаемое значение original_url (если wantShortURL)
	}{
		{
			name:         "валидный URL",
			body:         `{"url":"https://example.com/long/path"}`,
			wantStatus:   http.StatusCreated,
			wantShortURL: true,
			wantOriginal: "https://example.com/long/path",
		},
		{
			name:       "невалидный URL",
			body:       `{"url":"not-a-url"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "битый JSON",
			body:       `{"url":`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "пустой URL",
			body:       `{"url":""}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shortener := NewURLShortener()
			req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			shortener.ShortenHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("статус = %d, ожидали %d (body: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if !tt.wantShortURL {
				return
			}

			var resp shortenResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatalf("не удалось разобрать ответ: %v", err)
			}
			if resp.ShortURL == "" {
				t.Errorf("ожидали непустой short_url")
			}
			if resp.OriginalURL != tt.wantOriginal {
				t.Errorf("original_url = %q, ожидали %q", resp.OriginalURL, tt.wantOriginal)
			}
			if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q, ожидали application/json", ct)
			}
		})
	}
}

func TestRedirectHandler(t *testing.T) {
	shortener := NewURLShortener()
	mux := shortener.routes()

	const original = "https://example.com/long/path"
	shortID, err := shortener.Shorten(original)
	if err != nil {
		t.Fatalf("подготовка не удалась: %v", err)
	}

	tests := []struct {
		name         string
		path         string
		wantStatus   int
		wantLocation string
	}{
		{"существующий short_url", "/" + shortID, http.StatusFound, original},
		{"несуществующий short_url", "/nonexistent", http.StatusNotFound, ""},
		{"пустой путь (отсутствующий short_url)", "/", http.StatusNotFound, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("статус = %d, ожидали %d", rec.Code, tt.wantStatus)
			}
			if tt.wantLocation != "" {
				if loc := rec.Header().Get("Location"); loc != tt.wantLocation {
					t.Errorf("Location = %q, ожидали %q", loc, tt.wantLocation)
				}
			}
		})
	}
}

// TestRedirectHandler_EmptyShortID фиксирует поведение самого обработчика при
// пустом path-параметре short_url (минуя маршрутизацию): должен вернуться 404.
func TestRedirectHandler_EmptyShortID(t *testing.T) {
	shortener := NewURLShortener()

	// Запрос без значения для {short_url} => r.PathValue("short_url") == "".
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rec := httptest.NewRecorder()

	shortener.RedirectHandler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("статус = %d, ожидали %d", rec.Code, http.StatusNotFound)
	}
	if loc := rec.Header().Get("Location"); loc != "" {
		t.Errorf("Location = %q, ожидали пустой заголовок", loc)
	}
}
