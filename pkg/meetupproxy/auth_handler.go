package meetupproxy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/wire"
	"sgf-meetup-api/pkg/meetupproxy/meetupproxyconfig"
)

type AuthHandler interface {
	GetAccessToken(ctx context.Context) (string, error)
}

type MeetupHttpAuthHandlerConfig struct {
	URL          string
	UserID       string
	ClientKey    string
	SigningKeyID string
	PrivateKey   []byte
}

func NewMeetupAuthHandlerConfig(config *meetupproxyconfig.Config) MeetupHttpAuthHandlerConfig {
	return MeetupHttpAuthHandlerConfig{
		URL:          config.MeetupAuthURL,
		UserID:       config.MeetupUserID,
		ClientKey:    config.MeetupClientKey,
		SigningKeyID: config.MeetupSigningKeyID,
		PrivateKey:   config.MeetupPrivateKey,
	}
}

type MeetupHttpAuthHandler struct {
	lock       sync.Mutex
	token      *authToken
	config     MeetupHttpAuthHandlerConfig
	httpClient *http.Client
	logger     *slog.Logger
}

func NewMeetupHttpAuthHandler(
	config MeetupHttpAuthHandlerConfig,
	httpClient *http.Client,
	logger *slog.Logger,
) *MeetupHttpAuthHandler {
	return &MeetupHttpAuthHandler{
		config:     config,
		httpClient: httpClient,
		logger:     logger,
	}
}

type authToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	ExpiresAt    time.Time
	TokenType    string `json:"token_type"`
}

func (ah *MeetupHttpAuthHandler) GetAccessToken(ctx context.Context) (string, error) {
	ah.lock.Lock()
	defer ah.lock.Unlock()

	if ah.token == nil || ah.token.isExpiring(time.Now()) {
		ah.logger.Info("fetching new access token from meetup")
		newToken, err := ah.getNewAccessToken(ctx)
		if err != nil {
			ah.logger.Error("Error fetching token", "err", err)
			return "", err
		}

		ah.token = newToken
	}

	return ah.token.AccessToken, nil
}

func (ah *MeetupHttpAuthHandler) getNewAccessToken(ctx context.Context) (*authToken, error) {
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
		ah.config.URL,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")
	// Required by Meetup
	req.Header.Add("User-Agent", userAgent)

	resp, err := ah.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid status code when fetching token: %v", resp.StatusCode)
	}

	token, err := ah.parseAuthToken(resp.Body)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (ah *MeetupHttpAuthHandler) createSignedJWT() (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer:    ah.config.ClientKey,
		Subject:   ah.config.UserID,
		Audience:  jwt.ClaimStrings{"api.meetup.com"},
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
	}

	signedJwt := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedJwt.Header["kid"] = ah.config.SigningKeyID

	key, err := jwt.ParseRSAPrivateKeyFromPEM(ah.config.PrivateKey)
	if err != nil {
		return "", err
	}

	return signedJwt.SignedString(key)
}

func (t *authToken) isExpiring(offset time.Time) bool {
	return offset.After(t.ExpiresAt.Add(-30 * time.Second))
}

func (ah *MeetupHttpAuthHandler) parseAuthToken(r io.Reader) (*authToken, error) {
	var newToken authToken
	err := json.NewDecoder(r).Decode(&newToken)
	if err != nil {
		return nil, err
	}

	newToken.ExpiresAt = time.Now().Add(time.Duration(newToken.ExpiresIn) * time.Second)

	return &newToken, nil
}

var AuthHandlerProviders = wire.NewSet(
	wire.Bind(new(AuthHandler), new(*MeetupHttpAuthHandler)),
	NewMeetupAuthHandlerConfig,
	NewMeetupHttpAuthHandler,
)
