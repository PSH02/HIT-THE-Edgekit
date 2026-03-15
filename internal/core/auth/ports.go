package auth

type TokenStore interface {
	StoreRefreshToken(userID, token string) error
	RevokeRefreshToken(userID, token string) error
	IsRefreshTokenRevoked(userID, token string) (bool, error)
}
