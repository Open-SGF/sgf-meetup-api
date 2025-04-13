package meetupproxy

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"sgf-meetup-api/pkg/shared/logging"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestAuthHandler_GetAccessToken_InitialFetch(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(authToken{
			AccessToken: "test-token",
			ExpiresIn:   3600,
			TokenType:   "Bearer",
		})
	}))
	defer ts.Close()

	privateKey, _ := generatePrivateKey()
	ah := NewMeetupHttpAuthHandler(MeetupHttpAuthHandlerConfig{
		url:        ts.URL,
		privateKey: privateKey,
	}, &http.Client{}, logging.NewMockLogger())

	token, err := ah.GetAccessToken(context.Background())

	require.NoError(t, err)

	assert.Equal(t, "test-token", token)
}

func TestAuthHandler_GetAccessToken_ValidUserAgent(t *testing.T) {
	var capturedUserAgent string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserAgent = r.UserAgent()
		_ = json.NewEncoder(w).Encode(authToken{
			AccessToken: "test",
			ExpiresIn:   300,
		})
	}))
	defer ts.Close()

	privateKey, _ := generatePrivateKey()
	ah := NewMeetupHttpAuthHandler(MeetupHttpAuthHandlerConfig{
		url:        ts.URL,
		privateKey: privateKey,
	}, &http.Client{}, logging.NewMockLogger())

	_, err := ah.GetAccessToken(context.Background())

	require.NoError(t, err)

	assert.NotEmpty(t, capturedUserAgent)
	assert.Equal(t, userAgent, capturedUserAgent)
}

func TestAuthHandler_GetAccessToken_ExpiredToken(t *testing.T) {
	callCount := 0
	var mu sync.Mutex
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		callCount++
		mu.Unlock()

		_ = json.NewEncoder(w).Encode(authToken{
			AccessToken: fmt.Sprintf("token-%d", callCount),
			ExpiresIn:   0,
		})
	}))
	defer ts.Close()

	privateKey, _ := generatePrivateKey()
	ah := NewMeetupHttpAuthHandler(MeetupHttpAuthHandlerConfig{
		url:        ts.URL,
		privateKey: privateKey,
	}, &http.Client{}, logging.NewMockLogger())

	token, err := ah.GetAccessToken(context.Background())

	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	newToken, err := ah.GetAccessToken(context.Background())

	require.NoError(t, err)

	assert.NotEqual(t, token, newToken)
}

func TestAuthHandler_GetAccessTokenn_HTTPErrorHandling(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	privateKey, _ := generatePrivateKey()
	ah := NewMeetupHttpAuthHandler(MeetupHttpAuthHandlerConfig{
		url:        ts.URL,
		privateKey: privateKey,
	}, &http.Client{}, logging.NewMockLogger())

	_, err := ah.GetAccessToken(context.Background())
	require.Error(t, err)
	assert.ErrorContains(t, err, "invalid status code")
}

func TestAuthHandler_GetAccessToken_ConcurrentRequests(t *testing.T) {
	var callCount int
	var mu sync.Mutex

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		callCount++
		mu.Unlock()

		_ = json.NewEncoder(w).Encode(authToken{
			AccessToken: fmt.Sprintf("token-%d", callCount),
			ExpiresIn:   3600,
			TokenType:   "Bearer",
		})
	}))
	defer ts.Close()

	privateKey, _ := generatePrivateKey()
	ah := NewMeetupHttpAuthHandler(MeetupHttpAuthHandlerConfig{
		url:        ts.URL,
		privateKey: privateKey,
	}, &http.Client{}, logging.NewMockLogger())

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = ah.GetAccessToken(context.Background())
		}()
	}
	wg.Wait()

	assert.Equal(t, 1, callCount)
}

func TestAuthHandler_createSignedJWT_ValidClaims(t *testing.T) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	privateKeyBytes, _ := privateKeyToBytes(privateKey)
	ah := &MeetupHttpAuthHandler{
		config: MeetupHttpAuthHandlerConfig{
			clientKey:    "test-client",
			userID:       "user1",
			signingKeyID: "key1",
			privateKey:   privateKeyBytes,
		},
	}

	tokenString, err := ah.createSignedJWT()

	require.NoError(t, err)

	token, _ := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	})

	claims := token.Claims.(jwt.MapClaims)

	assert.Equal(t, "test-client", claims["iss"])
	assert.Equal(t, "user1", claims["sub"])
	assert.ElementsMatch(t, []string{"api.meetup.com"}, claims["aud"])
	assert.Equal(t, "key1", token.Header["kid"])
}

func TestAuthHandler_TestParseAuthToken(t *testing.T) {
	jsonResponse := `{
		"access_token": "parsed-token",
		"expires_in": 300,
		"token_type": "Bearer"
	}`

	ah := NewMeetupHttpAuthHandler(MeetupHttpAuthHandlerConfig{}, &http.Client{}, logging.NewMockLogger())

	token, err := ah.parseAuthToken(strings.NewReader(jsonResponse))

	require.NoError(t, err)

	expectedExpiry := time.Now().Add(300 * time.Second)
	assert.WithinRange(t, token.ExpiresAt, expectedExpiry.Add(-5*time.Second), expectedExpiry.Add(5*time.Second))
}

func TestIsExpiring(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		expiresAt time.Time
		checkTime time.Time
		expected  bool
	}{
		{
			name:      "valid_when_well_before_expiration",
			expiresAt: now.Add(35 * time.Second),
			checkTime: now,
			expected:  false,
		},
		{
			name:      "expiring_when_less_than_30s_remaining",
			expiresAt: now.Add(29 * time.Second),
			checkTime: now,
			expected:  true,
		},
		{
			name:      "exactly_30s_remaining",
			expiresAt: now.Add(30 * time.Second),
			checkTime: now,
			expected:  false,
		},
		{
			name:      "already_expired",
			expiresAt: now.Add(-1 * time.Second),
			checkTime: now,
			expected:  true,
		},
		{
			name:      "exact_expiration_threshold",
			expiresAt: now.Add(30 * time.Second),
			checkTime: now.Add(1 * time.Second),
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := &authToken{
				ExpiresAt: tt.expiresAt,
			}

			assert.Equal(t, tt.expected, token.isExpiring(tt.checkTime))
		})
	}
}

func privateKeyToBytes(key *rsa.PrivateKey) ([]byte, error) {
	privBytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, err
	}

	privPem := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privBytes,
	})

	return privPem, nil
}

func generatePrivateKey() ([]byte, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	return privateKeyToBytes(privateKey)
}
