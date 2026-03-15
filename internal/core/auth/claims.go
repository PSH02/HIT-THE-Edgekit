package auth

import "context"

type Subject struct {
	UserID string
	Role   string
	Scopes []string
}

type AuthContext struct {
	UserID string
	Role   string
	Scopes []string
}

type contextKey struct{}

func WithAuth(ctx context.Context, ac *AuthContext) context.Context {
	return context.WithValue(ctx, contextKey{}, ac)
}

func FromContext(ctx context.Context) (*AuthContext, bool) {
	ac, ok := ctx.Value(contextKey{}).(*AuthContext)
	return ac, ok
}

func (ac *AuthContext) HasScope(scope string) bool {
	for _, s := range ac.Scopes {
		if s == scope {
			return true
		}
	}
	return false
}

func (ac *AuthContext) HasRole(role string) bool {
	return ac.Role == role
}
