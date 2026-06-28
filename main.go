package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

// shortenRequest — тело запроса POST /shorten.
type shortenRequest struct {
	URL string `json:"url"`
}

// shortenResponse — тело ответа POST /shorten.
type shortenResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// errorResponse — единый формат тела ответа при ошибке.
type errorResponse struct {
	Error string `json:"error"`
}

// ShortenHandler обрабатывает POST /shorten: принимает JSON с url и
// возвращает сгенерированный короткий идентификатор.
func (us *URLShortener) ShortenHandler(w http.ResponseWriter, r *http.Request) {
	var req shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	shortID, err := us.Shorten(req.URL)
	if err != nil {
		if errors.Is(err, ErrInvalidURL) {
			writeError(w, http.StatusBadRequest, "invalid URL")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusCreated, shortenResponse{
		ShortURL:    shortID,
		OriginalURL: req.URL,
	})
}

// RedirectHandler обрабатывает GET /{short_url}: отдаёт 302-редирект на
// оригинальный URL либо 404, если идентификатор не найден.
func (us *URLShortener) RedirectHandler(w http.ResponseWriter, r *http.Request) {
	shortID := r.PathValue("short_url")

	originalURL, err := us.GetOriginal(shortID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "short URL not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}

// routes собирает маршруты сервиса в один ServeMux.
func (us *URLShortener) routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /shorten", us.ShortenHandler)
	mux.HandleFunc("GET /{short_url}", us.RedirectHandler)
	return mux
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}

func main() {
	shortener := NewURLShortener()

	addr := ":8080"
	log.Printf("URL shortener listening on %s", addr)
	if err := http.ListenAndServe(addr, shortener.routes()); err != nil {
		log.Fatal(err)
	}
}
