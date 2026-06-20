package response

import (
	"net/http"
	"testing"

	"github.com/d1vbyz3r0/typed/internal/parser/headers"
	"github.com/d1vbyz3r0/typed/internal/parser/response/codes"
	"github.com/d1vbyz3r0/typed/internal/parser/response/mime"
	"github.com/d1vbyz3r0/typed/internal/testsuite"
	"github.com/stretchr/testify/require"
)

func TestStatusCodeMapping_extractResponsesWebsockets(t *testing.T) {
	upgradeHeaders := []headers.Header{
		{
			Name:     "Connection",
			Type:     stringType,
			Required: true,
			Value:    "Upgrade",
		},
		{
			Name:     "Upgrade",
			Type:     stringType,
			Required: true,
			Value:    "websocket",
		},
	}

	tests := []struct {
		name        string
		wantHeaders []headers.Header
		handlerName string
	}{
		{
			name:        "xnet websocket",
			wantHeaders: upgradeHeaders,
			handlerName: "XNetWebsocket",
		},
		{
			name:        "gorilla websocket",
			wantHeaders: upgradeHeaders,
			handlerName: "GorillaWebsocket",
		},
		{
			name:        "coder websocket",
			wantHeaders: upgradeHeaders,
			handlerName: "CoderWebsocket",
		},
		{
			name:        "no websockets",
			wantHeaders: nil,
			handlerName: "Regular",
		},
	}

	cr, err := codes.NewResolver()
	require.NoError(t, err)
	mr, err := mime.NewResolver()
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkg, fun := testsuite.LoadFixtureFunc(t, "websockets", tt.handlerName)
			m := NewStatusCodeMapping(fun, cr, mr, pkg.TypesInfo)
			if len(tt.wantHeaders) == 0 {
				require.NotContains(t, m, http.StatusSwitchingProtocols)
				return
			}
			require.ElementsMatch(t, tt.wantHeaders, m[http.StatusSwitchingProtocols][0].Headers)
		})
	}
}

func TestXNetWebsocket(t *testing.T) {
	pkg, fun := testsuite.LoadFixtureFunc(t, "websockets", "XNetWebsocket")
	found := hasXNetWebSocketUsages(fun, pkg.TypesInfo)
	require.True(t, found)
}

func TestGorillaWebsocket(t *testing.T) {
	pkg, fun := testsuite.LoadFixtureFunc(t, "websockets", "GorillaWebsocket")
	found := hasGorillaWebSocketUsages(fun, pkg.TypesInfo)
	require.True(t, found)
}

func TestCoderWebsocket(t *testing.T) {
	pkg, fun := testsuite.LoadFixtureFunc(t, "websockets", "CoderWebsocket")
	found := hasCoderWebsocketUsages(fun, pkg.TypesInfo)
	require.True(t, found)
}
