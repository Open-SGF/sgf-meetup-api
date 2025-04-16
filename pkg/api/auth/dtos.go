package auth

type authRequestDTO struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

type refreshTokenRequestDTO struct {
	RefreshToken string `json:"refreshToken"`
}

type authResponseDTO struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int    `json:"expiresIn"`
}
