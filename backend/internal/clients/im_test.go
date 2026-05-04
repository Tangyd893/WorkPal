package clients

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Tangyd893/WorkPal/backend/internal/common/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIMClientSendsInternalToken(t *testing.T) {
	var gotToken string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotToken = r.Header.Get(middleware.InternalTokenHeader)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"message":"ok","data":{"is_member":true}}`))
	}))
	t.Cleanup(server.Close)

	client := NewIMClient(server.URL, "secret-token")
	isMember, err := client.IsMember(context.Background(), 1, 2)

	require.NoError(t, err)
	assert.True(t, isMember)
	assert.Equal(t, "secret-token", gotToken)
}
