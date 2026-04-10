package domain

type RegisterRequest struct {
	Email         string `json:"email"`
	Password      string `json:"password"`
	Name          string `json:"name"`
	HouseholdName string `json:"householdName,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
