package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/url"
	"sync"
)

// Sentinel-ошибки бизнес-логики.
var (
	// ErrInvalidURL возвращается, когда переданный URL невалиден.
	ErrInvalidURL = errors.New("invalid URL")
	// ErrNotFound возвращается, когда короткий идентификатор не найден.
	ErrNotFound = errors.New("short URL not found")
)

// shortIDLen — длина генерируемого короткого идентификатора (в диапазоне 6-8).
const shortIDLen = 8

// URLShortener хранит соответствие коротких идентификаторов и оригинальных URL.
type URLShortener struct {
	urls map[string]string
	mu   sync.RWMutex
}

// NewURLShortener создаёт готовый к работе экземпляр URLShortener.
func NewURLShortener() *URLShortener {
	return &URLShortener{
		urls: make(map[string]string),
	}
}

// Shorten создаёт короткий идентификатор для URL.
func (us *URLShortener) Shorten(originalURL string) (string, error) {
	if !isValidURL(originalURL) {
		return "", ErrInvalidURL
	}

	us.mu.Lock()
	defer us.mu.Unlock()

	// Подбираем уникальный идентификатор на случай коллизии.
	shortID := generateShortID()
	for {
		if _, exists := us.urls[shortID]; !exists {
			break
		}
		shortID = generateShortID()
	}

	us.urls[shortID] = originalURL
	return shortID, nil
}

// GetOriginal возвращает оригинальный URL по короткому идентификатору.
func (us *URLShortener) GetOriginal(shortID string) (string, error) {
	us.mu.RLock()
	defer us.mu.RUnlock()

	originalURL, ok := us.urls[shortID]
	if !ok {
		return "", ErrNotFound
	}
	return originalURL, nil
}

// generateShortID генерирует случайный короткий идентификатор длиной shortIDLen.
func generateShortID() string {
	// 6 случайных байт дают 8 символов base64 (RawURLEncoding, без паддинга).
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand на поддерживаемых платформах не возвращает ошибку;
		// паникуем, чтобы не выдать предсказуемый идентификатор.
		panic(err)
	}
	id := base64.RawURLEncoding.EncodeToString(b)
	if len(id) > shortIDLen {
		id = id[:shortIDLen]
	}
	return id
}

// isValidURL проверяет, что строка является корректным HTTP/HTTPS адресом.
func isValidURL(str string) bool {
	if str == "" {
		return false
	}
	u, err := url.Parse(str)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	return u.Host != ""
}
