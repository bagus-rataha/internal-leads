package dto

// RegisterInput for user registration
type RegisterInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=100"`
	Name     string `json:"name" validate:"required,min=2,max=100"`
}

// LoginInput for user login
type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RefreshTokenInput for token refresh.
// Optional in body — handler reads cookie first and falls back to this field.
type RefreshTokenInput struct {
	RefreshToken string `json:"refresh_token"`
}

// TokenResponse for auth responses
type TokenResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         UserResponse `json:"user"`
}
