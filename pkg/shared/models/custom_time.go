package models

import (
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var customTimeFormat = "2006-01-02T15:04-07:00"

var customTimeParseFormats = []string{time.RFC3339, customTimeFormat}

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
	var parsed time.Time
	var err error

	for _, layout := range customTimeParseFormats {
		parsed, err = time.Parse(layout, str)
		if err == nil {
			ct.Time = parsed
			return nil
		}
	}

	return err
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
