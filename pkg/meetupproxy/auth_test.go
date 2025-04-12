package meetupproxy

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sgf-meetup-api/pkg/logging"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGetAccessToken_InitialFetch(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(authToken{
			AccessToken: "test-token",
			ExpiresIn:   3600,
			TokenType:   "Bearer",
		})
	}))
	defer ts.Close()

	privateKey, _ := generatePrivateKey()
	ah := NewAuthHandler(AuthHandlerConfig{
		url:        ts.URL,
		privateKey: privateKey,
	}, &http.Client{}, logging.NewMockLogger())

	token, err := ah.GetAccessToken(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if token != "test-token" {
		t.Fatalf("Expected test-token, got %s", token)
	}
}

func TestAuthRequest_ValidUserAgent(t *testing.T) {
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
	ah := NewAuthHandler(AuthHandlerConfig{
		url:        ts.URL,
		privateKey: privateKey,
	}, &http.Client{}, logging.NewMockLogger())

	_, err := ah.GetAccessToken(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if capturedUserAgent == "" {
		t.Fatal("User-Agent header missing from request")
	}

	if capturedUserAgent != userAgent {
		t.Fatalf("Expected User-Agent %q, got %q",
			userAgent, capturedUserAgent)
	}
}

func TestGetAccessToken_ExpiredToken(t *testing.T) {
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
	ah := NewAuthHandler(AuthHandlerConfig{
		url:        ts.URL,
		privateKey: privateKey,
	}, &http.Client{}, logging.NewMockLogger())

	token, err := ah.GetAccessToken(context.Background())

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	newToken, err := ah.GetAccessToken(context.Background())

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if token == newToken {
		t.Fatalf("Expected newToken to be different from original token")
	}
}

func TestGetNewAccessToken_HTTPErrorHandling(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	privateKey, _ := generatePrivateKey()
	ah := NewAuthHandler(AuthHandlerConfig{
		url:        ts.URL,
		privateKey: privateKey,
	}, &http.Client{}, logging.NewMockLogger())

	_, err := ah.GetAccessToken(context.Background())
	if err == nil || !strings.Contains(err.Error(), "invalid status code") {
		t.Fatalf("Expected status code error, got: %v", err)
	}
}

func TestCreateSignedJWT_ValidClaims(t *testing.T) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	privateKeyBytes, _ := privateKeyToBytes(privateKey)
	ah := &authHandler{
		config: AuthHandlerConfig{
			clientKey:    "test-client",
			userID:       "user1",
			signingKeyID: "key1",
			privateKey:   privateKeyBytes,
		},
	}

	tokenString, err := ah.createSignedJWT()
	if err != nil {
		t.Fatalf("JWT creation failed: %v", err)
	}

	token, _ := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	})

	claims := token.Claims.(jwt.MapClaims)
	if claims["iss"] != "test-client" {
		t.Fatalf("Invalid issuer claim: %v", claims["iss"])
	}

	if claims["sub"] != "user1" {
		t.Fatalf("Invalid subject claim: %v", claims["sub"])
	}

	if reflect.DeepEqual(claims["aud"], []string{"api.meetup.com"}) {
		t.Fatalf("Invalid audience claim: %v", claims["aud"])
	}

	if token.Header["kid"] != "key1" {
		t.Fatalf("Invalid kid header: %v", token.Header["kid"])
	}
}

func TestParseAuthToken(t *testing.T) {
	jsonResponse := `{
		"access_token": "parsed-token",
		"expires_in": 300,
		"token_type": "Bearer"
	}`

	token, err := parseAuthToken(strings.NewReader(jsonResponse))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expectedExpiry := time.Now().Add(300 * time.Second)
	if token.ExpiresAt.Before(expectedExpiry.Add(-5*time.Second)) || token.ExpiresAt.After(expectedExpiry.Add(5*time.Second)) {
		t.Errorf("ExpiresAt not within expected range: %v", token.ExpiresAt)
	}
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

			if actual := token.isExpiring(tt.checkTime); actual != tt.expected {
				t.Errorf("For expiresAt %v and checkTime %v\nExpected %v, got %v",
					token.ExpiresAt, tt.checkTime, tt.expected, actual)
			}
		})
	}
}

func TestGetAccessToken_ConcurrentRequests(t *testing.T) {
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
	ah := NewAuthHandler(AuthHandlerConfig{
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

	if callCount != 1 {
		t.Fatalf("Expected 1 token fetch, got %d", callCount)
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
