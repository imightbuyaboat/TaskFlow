package auth

type TokenManager interface {
	CreateToken(userID uint64) (string, error)
	ValidateToken(tokenStr string) (uint64, error)
}
