package models

type AuthResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}
