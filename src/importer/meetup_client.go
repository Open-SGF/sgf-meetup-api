package importer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sgf-meetup-api/src/constants"
	"sgf-meetup-api/src/models"
	"strings"
)

type MeetupClient struct {
	ctx   context.Context
	url   string
	token string
}

func NewMeetupClient(ctx context.Context, url, token string) *MeetupClient {
	return &MeetupClient{
		ctx:   ctx,
		token: token,
		url:   url,
	}
}

const getFutureEventsQuery = `
  query ($urlname: String!, $itemsNum: Int!, $cursor: String) {
	events: groupByUrlname(urlname: $urlname) {
	  unifiedEvents(input: { first: $itemsNum, after: $cursor }) {
		count
		pageInfo {
		  endCursor
		  hasNextPage
		}
		edges {
		  node {
			id
			title
			eventUrl
			description
			dateTime
			duration
			venue {
			  name
			  address
			  city
			  state
			  postalCode
			}
			group {
			  name
			  urlname
			}
			host {
			  name
			}
			images {
			  baseUrl
			  preview
			}
		  }
		}
	  }
	}
  }
`

type MeetupFutureEventsResponse struct {
	Data struct {
		Events struct {
			UnifiedEvents struct {
				Count    int `json:"count"`
				PageInfo struct {
					EndCursor   string `json:"endCursor"`
					HasNextPage bool   `json:"hasNextPage"`
				} `json:"pageInfo"`
				Edge struct {
					Node models.MeetupEvent `json:"node"`
				} `json:"edge"`
			} `json:"unifiedEvents"`
		} `json:"events"`
	} `json:"data"`
}

func (client *MeetupClient) GetFutureEventsForGroup(group string, cursor string) (*MeetupFutureEventsResponse, error) {
	queryVars := map[string]interface{}{
		"group":    group,
		"itemsNum": 10,
	}

	if cursor != "" {
		queryVars["cursor"] = cursor
	}

	body, err := client.MakeGraphQlRequest(getFutureEventsQuery, queryVars)

	if err != nil {
		return nil, err
	}

	defer func() { _ = body.Close() }()

	var response MeetupFutureEventsResponse
	jsonErr := json.NewDecoder(body).Decode(&response)

	if jsonErr != nil {
		return nil, jsonErr
	}

	return &response, nil
}

func (client *MeetupClient) MakeGraphQlRequest(query string, variables map[string]interface{}) (io.ReadCloser, error) {
	type Body struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}

	reqBody := Body{
		Query:     query,
		Variables: variables,
	}

	reqBodyJson, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(client.ctx, http.MethodPost, client.url, strings.NewReader(string(reqBodyJson)))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", constants.UserAgent)
	req.Header.Add("Authorization", "Bearer "+client.token)

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("expected status code 200, got %v", resp.StatusCode)
	}

	return resp.Body, nil
}
