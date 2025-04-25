package models

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"strings"
	"time"
)

var customTimeFormat = "2006-01-02T15:04-07:00"

type CustomTime struct {
	time.Time
}

//goland:noinspection GoMixedReceiverTypes
func (ct *CustomTime) String() string {
	return ct.Format(customTimeFormat)
}

//goland:noinspection GoMixedReceiverTypes
func (ct *CustomTime) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), `"`) // Remove JSON quotes
	parsed, err := time.Parse(customTimeFormat, str)
	if err != nil {
		return err
	}
	ct.Time = parsed
	return nil
}

//goland:noinspection GoMixedReceiverTypes
func (ct CustomTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + ct.String() + `"`), nil
}

//goland:noinspection GoMixedReceiverTypes
func (ct CustomTime) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	return &types.AttributeValueMemberS{
		Value: ct.Format(time.RFC3339),
	}, nil
}

//goland:noinspection GoMixedReceiverTypes
func (ct *CustomTime) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	s, ok := av.(*types.AttributeValueMemberS)
	if !ok {
		return nil
	}

	parsed, err := time.Parse(time.RFC3339, s.Value)
	if err != nil {
		return err
	}
	ct.Time = parsed
	return nil
}
