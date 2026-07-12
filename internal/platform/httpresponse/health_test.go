package httpresponse

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHealth(t *testing.T) {
	t.Run("it should return the application health", func(t *testing.T) {
		gin.SetMode(gin.TestMode)

		responseRecorder := httptest.NewRecorder()
		context, _ := gin.CreateTestContext(responseRecorder)

		Health(context)

		if responseRecorder.Code != http.StatusOK {
			t.Fatalf("expected status %d, received %d", http.StatusOK, responseRecorder.Code)
		}

		expectedBody := "{\"status\":\"UP\"}"
		if responseRecorder.Body.String() != expectedBody {
			t.Fatalf("expected body %s, received %s", expectedBody, responseRecorder.Body.String())
		}
	})
}
