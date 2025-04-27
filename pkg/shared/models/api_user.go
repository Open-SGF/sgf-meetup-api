package models

type APIUser struct {
	ClientID           string `dynamodbav:"clientId"`
	HashedClientSecret []byte `dynamodbav:"hashedClientSecret"`
}
