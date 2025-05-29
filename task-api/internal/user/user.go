package user

type User struct {
	Email    string `json:"login" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}
