package dto

type CredentialsRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateCustomerRequest struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	Password     string `json:"password"`
	PasswordHash string `json:"passwordHash"`
}
