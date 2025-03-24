package meetupproxy

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type AuthHandler interface {
	GetAccessToken(ctx context.Context) (string, error)
}

type authHandler struct {
	lock   sync.Mutex
	token  *authToken
	config AuthHandlerConfig
}

type AuthHandlerConfig struct {
	url          string
	userID       string
	clientKey    string
	signingKeyID string
	privateKey   []byte
}

func NewAuthHandler(c AuthHandlerConfig) AuthHandler {
	return &authHandler{
		config: c,
	}
}

func NewAuthHandlerFromConfig(config *Config) AuthHandler {
	return NewAuthHandler(AuthHandlerConfig{
		url:          config.MeetupAuthURL,
		userID:       config.MeetupUserID,
		clientKey:    config.MeetupClientKey,
		signingKeyID: config.MeetupSigningKeyID,
		privateKey:   config.MeetupPrivateKey,
	})
}

type authToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	ExpiresAt    time.Time
	TokenType    string `json:"token_type"`
}

func (ah *authHandler) GetAccessToken(ctx context.Context) (string, error) {
	ah.lock.Lock()
	defer ah.lock.Unlock()

	if ah.token == nil || ah.token.isExpiring(time.Now()) {
		newToken, err := ah.getNewAccessToken(ctx)

		if err != nil {
			return "", err
		}

		ah.token = newToken
	}

	return ah.token.AccessToken, nil
}

func (ah *authHandler) getNewAccessToken(ctx context.Context) (*authToken, error) {
	signedJwt, err := ah.createSignedJWT()

	if err != nil {
		return nil, err
	}

	form := url.Values{}
	form.Add("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	form.Add("assertion", signedJwt)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		ah.config.url,
		strings.NewReader(form.Encode()),
	)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")
	// Required by Meetup
	req.Header.Add("User-Agent", userAgent)

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid status code when fetching token")
	}

	token, err := parseAuthToken(resp.Body)

	if err != nil {
		return nil, err
	}

	return token, nil
}

func (ah *authHandler) createSignedJWT() (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer:    ah.config.clientKey,
		Subject:   ah.config.userID,
		Audience:  jwt.ClaimStrings{"api.meetup.com"},
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
	}

	signedJwt := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedJwt.Header["kid"] = ah.config.signingKeyID

	key, err := jwt.ParseRSAPrivateKeyFromPEM(ah.config.privateKey)

	if err != nil {
		return "", err
	}

	return signedJwt.SignedString(key)
}

func (t *authToken) isExpiring(offset time.Time) bool {
	return offset.After(t.ExpiresAt.Add(-30 * time.Second))
}

func parseAuthToken(r io.Reader) (*authToken, error) {
	var newToken authToken
	err := json.NewDecoder(r).Decode(&newToken)
	if err != nil {
		return nil, err
	}

	newToken.ExpiresAt = time.Now().Add(time.Duration(newToken.ExpiresIn) * time.Second)

	return &newToken, nil
}
