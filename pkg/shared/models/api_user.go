package models

type APIUser struct {
	ClientID           string `dynamodbav:"clientId"`
	HashedClientSecret string `dynamodbav:"hashedClientSecret"`
}
