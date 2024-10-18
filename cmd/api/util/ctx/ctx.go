package ctx

import "context"

const requestIdKey strKey = "request.id"

type strKey string

func RequestId(ctx context.Context) string {
	requestId, _ := ctx.Value(requestIdKey).(string)

	return requestId
}

func SetRequestId(ctx context.Context, requestId string) context.Context {
	return context.WithValue(ctx, requestIdKey, requestId)
}
