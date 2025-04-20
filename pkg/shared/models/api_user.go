package models

type ApiUser struct {
	ClientID           string `dynamodbav:"clientId"`
	HashedClientSecret string `dynamodbav:"hashedClientSecret"`
}
