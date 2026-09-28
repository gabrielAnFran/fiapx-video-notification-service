//go:build integration

package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gabrielAnFran/fiapx-video-notification-service/internal/presentation/handlers"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestHealthHandler_Readyz_DBAvailable covers Readyz's happy path against
// the real shared Postgres container — health_handler_test.go (plain unit
// test, no build tag) already covers Healthz and Readyz's unavailable
// branch without needing real infra.
func TestHealthHandler_Readyz_DBAvailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newTestDB(t)
	h := handlers.NewHealthHandler(db)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/readyz", nil)

	h.Readyz(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status":"ok"}`, w.Body.String())
}
