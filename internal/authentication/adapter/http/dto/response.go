package dto

type UserResponse struct {
	ID        string   `json:"id"`
	Email     string   `json:"email"`
	Roles     []string `json:"roles"`
	CreatedAt string   `json:"created_at"`
}

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
}
