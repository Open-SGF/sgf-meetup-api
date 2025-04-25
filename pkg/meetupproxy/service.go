package meetupproxy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sgf-meetup-api/pkg/meetupproxy/meetupproxyconfig"
	"strings"
)

const userAgent = "curl/8.7.1"

type ServiceConfig struct {
	URL string
}

func NewServiceConfig(config *meetupproxyconfig.Config) ServiceConfig {
	return ServiceConfig{
		URL: config.MeetupAPIURL,
	}
}

type Service struct {
	config     ServiceConfig
	logger     *slog.Logger
	httpClient *http.Client
	auth       AuthHandler
}

func NewService(config ServiceConfig, httpClient *http.Client, auth AuthHandler, logger *slog.Logger) *Service {
	return &Service{
		config:     config,
		httpClient: httpClient,
		auth:       auth,
		logger:     logger,
	}
}

type Request struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

type Response map[string]any

func (s *Service) HandleRequest(ctx context.Context, req Request) (*Response, error) {
	token, err := s.auth.GetAccessToken(ctx)

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

	meetupReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.config.URL, strings.NewReader(string(reqBodyJson)))

	if err != nil {
		return nil, err
	}

	meetupReq.Header.Add("Content-Type", "application/json")
	meetupReq.Header.Add("Accept", "application/json")
	meetupReq.Header.Add("User-Agent", userAgent)
	meetupReq.Header.Add("Authorization", "Bearer "+token)

	meetupResp, err := s.httpClient.Do(meetupReq)

	if err != nil {
		return nil, err
	}

	defer func() { _ = meetupResp.Body.Close() }()

	if meetupResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(meetupResp.Body)
		s.logger.Error("Error fetching data from meetup", "statusCode", meetupResp.StatusCode, "body", string(body))
		return nil, fmt.Errorf("expected status code 200, got %v", meetupResp.StatusCode)
	}

	var resp Response
	err = json.NewDecoder(meetupResp.Body).Decode(&resp)

	if err != nil {
		return nil, err
	}

	return &resp, nil
}
