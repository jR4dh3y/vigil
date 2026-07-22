package auth

import "context"

type contextKey int

const userContextKey contextKey = 1

// User is the authenticated principal attached to a request context.
type User struct {
	ID       string
	Username string
	Role     string
}

// WithUser returns a child context carrying u.
func WithUser(ctx context.Context, u *User) context.Context {
	return context.WithValue(ctx, userContextKey, u)
}

// UserFromContext returns the authenticated user, or nil if unauthenticated.
func UserFromContext(ctx context.Context) *User {
	u, _ := ctx.Value(userContextKey).(*User)
	return u
}
