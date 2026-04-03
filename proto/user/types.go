package userpb

// RegisterRequest is the payload for creating a new user.
type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterResponse is returned after successful registration.
type RegisterResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

// LoginRequest is the payload for user authentication.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse contains the JWT and user ID on success.
type LoginResponse struct {
	Token  string `json:"token"`
	UserID string `json:"user_id"`
}

// GetUserRequest looks up a user by ID.
type GetUserRequest struct {
	ID string `json:"id"`
}

// GetUserResponse contains the user's public profile.
type GetUserResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
