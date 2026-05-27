package models

type RegisterRequest struct {
	Login    string `json:"login" binding:"required,min=3,max=100"`
	Name     string `json:"name" binding:"required,min=1,max=200"`
	Password string `json:"password" binding:"required,min=6,max=72"`
}
