package meetupproxy

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sgf-meetup-api/pkg/logging"
	"strings"
)

const userAgent = "curl/8.7.1"

type Proxy struct {
	url    string
	logger *slog.Logger
	auth   AuthHandler
}

func New(url string, logger *slog.Logger, auth AuthHandler) *Proxy {
	return &Proxy{
		url:    url,
		logger: logger,
		auth:   auth,
	}
}

func NewFromConfig(config *Config) *Proxy {
	logger := logging.DefaultLogger(config.LogLevel)
	auth := NewAuthHandler(AuthHandlerConfig{
		url:          config.MeetupAuthURL,
		userID:       config.MeetupUserID,
		clientKey:    config.MeetupClientKey,
		signingKeyID: config.MeetupSigningKeyID,
		privateKey:   config.MeetupPrivateKey,
	}, logger)
	return New(config.MeetupAPIURL, logger, auth)
}

type Request struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

type Response map[string]any

func (p *Proxy) HandleRequest(ctx context.Context, req Request) (*Response, error) {
	token, err := p.auth.GetAccessToken(ctx)

	if err != nil {
		return nil, err
	}

	if req.Query == "" {
		return nil, err
	}

	reqBodyJson, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	meetupReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.url, strings.NewReader(string(reqBodyJson)))

	if err != nil {
		return nil, err
	}

	meetupReq.Header.Add("Content-Type", "application/json")
	meetupReq.Header.Add("Accept", "application/json")
	meetupReq.Header.Add("User-Agent", userAgent)
	meetupReq.Header.Add("Authorization", "Bearer "+token)

	httpClient := &http.Client{Transport: logging.NewHttpLoggingTransport(p.logger)}
	meetupResp, err := httpClient.Do(meetupReq)

	if err != nil {
		return nil, err
	}

	defer func() { _ = meetupResp.Body.Close() }()

	if meetupResp.StatusCode != http.StatusOK {
		p.logger.Error("Error fetching data from meetup", "statusCode", meetupResp.StatusCode)
		return nil, fmt.Errorf("expected status code 200, got %v", meetupResp.StatusCode)
	}

	var resp Response
	err = json.NewDecoder(meetupResp.Body).Decode(&resp)

	if err != nil {
		return nil, err
	}

	return &resp, nil

}
