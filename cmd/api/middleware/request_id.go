package middleware

import (
	"net/http"

	"github.com/rs/xid"

	ctxUtil "github.com/bartosz121/webhooks-api/cmd/api/util/ctx"
)

const requestIdHeader = "x-request-id"

func RequestId(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		requestId := r.Header.Get(requestIdHeader)
		if requestId == "" {
			requestId = xid.New().String()
		}

		ctx = ctxUtil.SetRequestId(ctx, requestId)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
