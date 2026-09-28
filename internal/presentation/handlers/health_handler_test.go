package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestHealthHandler_Healthz(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHealthHandler(nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/healthz", nil)

	h.Healthz(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status":"ok"}`, w.Body.String())
}

func TestHealthHandler_Readyz_DBUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// gorm.Open connects eagerly; against an unreachable port it fails fast
	// (connection refused) and returns a non-nil *gorm.DB whose underlying
	// *sql.DB is already closed, so Ping always errors — exactly the
	// "database unavailable" case Readyz needs to report.
	db, _ := gorm.Open(postgres.Open("host=127.0.0.1 port=1 user=x password=x dbname=x sslmode=disable"), &gorm.Config{})
	require.NotNil(t, db)
	h := NewHealthHandler(db)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/readyz", nil)

	h.Readyz(c)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.JSONEq(t, `{"status":"unavailable"}`, w.Body.String())
}
