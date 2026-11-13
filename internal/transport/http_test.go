package transport

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireBearer(t *testing.T) {
	h := requireBearer("s3cret", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	for header, want := range map[string]int{
		"":              http.StatusUnauthorized,
		"Bearer":        http.StatusUnauthorized,
		"Bearer nope":   http.StatusUnauthorized,
		"Basic s3cret":  http.StatusUnauthorized,
		"Bearer s3cret": http.StatusOK,
		"bearer s3cret": http.StatusOK,
	} {
		req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
		req.Header.Set("Authorization", header)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != want {
			t.Errorf("Authorization %q: status %d, want %d", header, rec.Code, want)
		}
	}
}
