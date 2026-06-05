package request

// Register -.
type Register struct {
	Username string `json:"username" validate:"required,min=3,max=255"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

// Login -.
type Login struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,max=72"`
}
