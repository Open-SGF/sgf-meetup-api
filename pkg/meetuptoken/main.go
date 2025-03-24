package meetuptoken

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"net/url"
	"sgf-meetup-api/pkg/constants"
	"strings"
	"time"
)

type meetupResponse struct {
	AccessToken string `json:"access_token"`
}

func GetToken(ctx context.Context, config *Config, clientId string) (string, error) {
	signedJwt, err := createSignedJwt(config)

	if err != nil {
		return "", err
	}

	form := url.Values{}
	form.Add("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	form.Add("assertion", signedJwt)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		config.MeetupAuthUrl,
		strings.NewReader(form.Encode()),
	)

	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")
	// Required by Meetup
	req.Header.Add("User-Agent", constants.UserAgent)

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("invalid status code when fetching token")
	}

	var result meetupResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", nil
	}

	return result.AccessToken, nil
}

func createSignedJwt(config *Config) (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer:    config.MeetupClientKey,
		Subject:   config.MeetupUserId,
		Audience:  jwt.ClaimStrings{"api.meetup.com"},
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = config.MeetupSigningKeyId

	key, err := jwt.ParseRSAPrivateKeyFromPEM(config.MeetupPrivateKey)

	if err != nil {
		return "", err
	}

	return token.SignedString(key)
}
