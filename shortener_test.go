package main

import (
	"errors"
	"testing"
)

func TestURLShortener_Shorten(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"валидный HTTP URL", "http://example.com", false},
		{"валидный HTTPS URL", "https://google.com/search?q=test", false},
		{"невалидный URL", "not-a-url", true},
		{"пустая строка", "", true},
		{"неподдерживаемая схема", "ftp://example.com", true},
	}

	shortener := NewURLShortener()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shortID, err := shortener.Shorten(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ошибка = %v, ожидали ошибку = %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if !errors.Is(err, ErrInvalidURL) {
					t.Errorf("ожидали ErrInvalidURL, получили %v", err)
				}
				return
			}
			if l := len(shortID); l < 6 || l > 8 {
				t.Errorf("длина короткого ID вне диапазона 6-8: %q (%d)", shortID, l)
			}
		})
	}
}

func TestURLShortener_Shorten_Unique(t *testing.T) {
	shortener := NewURLShortener()
	const n = 100
	seen := make(map[string]struct{}, n)
	for i := 0; i < n; i++ {
		id, err := shortener.Shorten("https://example.com")
		if err != nil {
			t.Fatalf("неожиданная ошибка: %v", err)
		}
		if _, dup := seen[id]; dup {
			t.Fatalf("сгенерирован дублирующийся идентификатор: %s", id)
		}
		seen[id] = struct{}{}
	}
}

func TestURLShortener_GetOriginal(t *testing.T) {
	shortener := NewURLShortener()
	const original = "https://example.com/long/path"
	shortID, err := shortener.Shorten(original)
	if err != nil {
		t.Fatalf("подготовка не удалась: %v", err)
	}

	tests := []struct {
		name     string
		shortID  string
		want     string
		wantErr  bool
		errIsNot bool // true => ожидаем ErrNotFound
	}{
		{"существующий ID", shortID, original, false, false},
		{"отсутствующий ID", "missing", "", true, true},
		{"пустой ID", "", "", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := shortener.GetOriginal(tt.shortID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ошибка = %v, ожидали ошибку = %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if tt.errIsNot && !errors.Is(err, ErrNotFound) {
					t.Errorf("ожидали ErrNotFound, получили %v", err)
				}
				return
			}
			if got != tt.want {
				t.Errorf("GetOriginal() = %q, ожидали %q", got, tt.want)
			}
		})
	}
}

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{"http", "http://example.com", true},
		{"https с путём и query", "https://google.com/search?q=test", true},
		{"без схемы", "example.com/path", false},
		{"схема ftp", "ftp://example.com", false},
		{"пустая строка", "", false},
		{"мусор", "not-a-url", false},
		{"http без хоста", "http://", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidURL(tt.url); got != tt.want {
				t.Errorf("isValidURL(%q) = %v, ожидали %v", tt.url, got, tt.want)
			}
		})
	}
}

func TestGenerateShortID(t *testing.T) {
	t.Run("длина в диапазоне 6-8", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			id := generateShortID()
			if l := len(id); l < 6 || l > 8 {
				t.Fatalf("длина %q вне диапазона 6-8: %d", id, l)
			}
		}
	})

	t.Run("значения различаются", func(t *testing.T) {
		a, b := generateShortID(), generateShortID()
		if a == b {
			t.Errorf("два вызова вернули одинаковое значение: %s", a)
		}
	})
}
