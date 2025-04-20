package auth

import "time"

type authRequestDTO struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

type refreshTokenRequestDTO struct {
	RefreshToken string `json:"refreshToken"`
}

type authResponseDTO struct {
	AccessToken           string    `json:"accessToken"`
	AccessTokenExpiresAt  time.Time `json:"accessTokenExpiresAt"`
	RefreshToken          string    `json:"refreshToken"`
	RefreshTokenExpiresAt time.Time `json:"refreshTokenExpiresAt"`
}
