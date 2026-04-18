package api

import (
	"context"

	"api-proxy/internal/model"
)

// Unexported key type used for attaching values to context.Context;
// keeps API-local keys from colliding with keys from other packages.
type ctxKey string

const userCtxKey ctxKey = "user"

// Derives a new context with the authenticated user attached under
// the package-private userCtxKey.
func setUserCtx(ctx context.Context, user *model.User) context.Context {
	return context.WithValue(ctx, userCtxKey, user)
}

// Retrieves the authenticated user previously stored via setUserCtx;
// returns nil when no user is present in the context.
func getUserCtx(ctx context.Context) *model.User {
	u, _ := ctx.Value(userCtxKey).(*model.User)
	return u
}
