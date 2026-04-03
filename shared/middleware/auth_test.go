package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "unit-test-secret"

func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AuthMiddleware(testSecret))
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"user_id": c.GetString("user_id"),
			"email":   c.GetString("email"),
		})
	})
	return r
}

func makeToken(userID, email, secret string, expiry time.Duration) string {
	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := tok.SignedString([]byte(secret))
	return signed
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	r := newTestRouter()
	token := makeToken("user-123", "alice@example.com", testSecret, time.Hour)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	r := newTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	r := newTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer not.a.real.token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	r := newTestRouter()
	token := makeToken("user-123", "alice@example.com", testSecret, -time.Hour) // expired 1h ago

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_WrongSecret(t *testing.T) {
	r := newTestRouter()
	token := makeToken("user-123", "alice@example.com", "different-secret", time.Hour)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_MalformedBearerScheme(t *testing.T) {
	r := newTestRouter()
	token := makeToken("user-123", "alice@example.com", testSecret, time.Hour)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Token "+token) // wrong scheme
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestGenerateToken_ValidAndParseable(t *testing.T) {
	tokenStr, err := GenerateToken("user-abc", "test@test.com", testSecret)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	claims := &Claims{}
	tok, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(testSecret), nil
	})
	if err != nil || !tok.Valid {
		t.Fatalf("token not valid: %v", err)
	}
	if claims.UserID != "user-abc" {
		t.Errorf("UserID = %q, want %q", claims.UserID, "user-abc")
	}
}
