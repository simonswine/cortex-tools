package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

type request struct {
	method    string
	namespace string
	name      string
}

func TestCortexClient_X(t *testing.T) {
	requestCh := make(chan request, 1)

	router := mux.NewRouter()
	router.Path("/api/v1/rules/{namespace}/{groupName}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		var (
			req request
			err error
		)

		for key, value := range vars {
			if key == "namespace" {
				req.namespace, err = url.PathUnescape(value)
				require.NoError(t, err)
			}

			if key == "groupName" {
				req.name, err = url.PathUnescape(value)
				require.NoError(t, err)
			}
		}

		req.method = r.Method
		requestCh <- req
		fmt.Fprintln(w, "hello")
	})

	ts := httptest.NewServer(router)
	defer ts.Close()

	client, err := New(Config{
		Address: ts.URL,
		ID:      "my-id",
		Key:     "my-key",
	})
	require.NoError(t, err)

	for _, tc := range []struct {
		test      string
		namespace string
		name      string
		method    string
	}{{
		test:      "regular-characters",
		namespace: "my-namespace",
		name:      "my-name",
		method:    "DELETE",
	}, {
		test:      "special-characters",
		namespace: "My: Namespace",
		name:      "My: Name",
		method:    "DELETE",
	}} {
		t.Run(tc.test, func(t *testing.T) {
			ctx := context.Background()
			require.NoError(t, client.DeleteRuleGroup(ctx, tc.namespace, tc.name))

			n := <-requestCh
			require.Equal(t, tc.namespace, n.namespace)
			require.Equal(t, tc.name, n.name)
			require.Equal(t, tc.method, n.method)

		})
	}

}
