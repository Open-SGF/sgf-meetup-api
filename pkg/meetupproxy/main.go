package meetupproxy

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const userAgent = "curl/8.7.1"

type Proxy struct {
	url  string
	auth AuthHandler
}

func New(url string, auth AuthHandler) *Proxy {
	return &Proxy{
		url:  url,
		auth: auth,
	}
}

func NewFromConfig(config *Config) *Proxy {
	auth := NewAuthHandlerFromConfig(config)
	return New(config.MeetupAPIURL, auth)
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

	httpClient := &http.Client{}
	meetupResp, err := httpClient.Do(meetupReq)

	if err != nil {
		return nil, err
	}

	defer func() { _ = meetupResp.Body.Close() }()

	if meetupResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expected status code 200, got %v", meetupResp.StatusCode)
	}

	var resp Response
	err = json.NewDecoder(meetupResp.Body).Decode(&resp)

	if err != nil {
		return nil, err
	}

	return &resp, nil

}
