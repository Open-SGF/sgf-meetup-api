package importer

import (
	"context"
	"encoding/json"
	"sgf-meetup-api/pkg/models"
)

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

type meetupRepository struct {
	handler GraphQLHandler
}

type MeetupRepository interface {
}

type GraphQLHandler interface {
	ExecuteQuery(ctx context.Context, query string, variables map[string]any) ([]byte, error)
}

func NewMeetupRepository(handler GraphQLHandler) MeetupRepository {
	return &meetupRepository{
		handler: handler,
	}
}

func executeGraphQLQuery[T any](ctx context.Context, c *meetupRepository, query string, variables map[string]any) (*T, error) {
	responseBytes, err := c.handler.ExecuteQuery(ctx, query, variables)

	if err != nil {
		return nil, err
	}

	var response T
	if err = json.Unmarshal(responseBytes, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

//func (s *Service) GetEventsForGroup(ctx context.Context, group string) ([]models.MeetupEvent, error) {
//	//fetchCutOff := time.Now().AddDate(0, 6, 0)
//	//var newestDate time.Time
//
//	//var response :=
//}
