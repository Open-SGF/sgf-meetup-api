package models

import (
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomTime_String(t *testing.T) {
	ct := CustomTime{
		time.Date(2025, 4, 25, 2, 19, 0, 0, time.UTC),
	}
	assert.Equal(t, "2025-04-25T02:19+00:00", ct.String())
}

func TestCustomTime_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectErr   bool
		expectedStr string
	}{
		{
			name:        "valid UTC time",
			input:       `"2025-04-25T02:19+00:00"`,
			expectedStr: "2025-04-25T02:19+00:00",
		},
		{
			name:        "valid time with offset",
			input:       `"2025-04-24T22:19-04:00"`,
			expectedStr: "2025-04-24T22:19-04:00",
		},
		{
			name:      "invalid format missing T",
			input:     `"2025-04-25 02:19+00:00"`,
			expectErr: true,
		},
		{
			name:      "invalid non-string input",
			input:     `12345`,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ct CustomTime
			err := json.Unmarshal([]byte(tt.input), &ct)

			if tt.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStr, ct.String())
		})
	}

	t.Run("pointer on struct", func(t *testing.T) {
		var someStruct struct {
			Time *CustomTime `json:"time"`
		}

		err := json.Unmarshal([]byte("{\"time\": \"2025-05-15T17:30-05:00\"}"), &someStruct)
		require.NoError(t, err)
		assert.Equal(t, "2025-05-15T17:30-05:00", someStruct.Time.String())
	})
}

func TestCustomTime_MarshalJSON(t *testing.T) {
	t.Run("successful marshal", func(t *testing.T) {
		ct := CustomTime{
			time.Date(2025, 4, 25, 2, 19, 0, 0, time.FixedZone("TEST", -4*60*60)),
		}

		data, err := json.Marshal(ct)
		require.NoError(t, err)
		assert.Equal(t, `"2025-04-25T02:19-04:00"`, string(data))
	})

	t.Run("zero time handling", func(t *testing.T) {
		var ct CustomTime // Zero time
		data, err := json.Marshal(ct)
		require.NoError(t, err)
		assert.Equal(t, `"0001-01-01T00:00+00:00"`, string(data))
	})
}

func TestCustomTime_DynamoDBAttributeMarshalUnmarshal(t *testing.T) {
	t.Run("marshal valid time", func(t *testing.T) {
		ct := CustomTime{
			Time: time.Date(2025, 4, 25, 2, 19, 0, 0, time.FixedZone("EDT", -4*60*60)),
		}

		av, err := ct.MarshalDynamoDBAttributeValue()
		require.NoError(t, err)

		s, ok := av.(*types.AttributeValueMemberS)
		require.True(t, ok, "Should return AttributeValueMemberS")
		assert.Equal(t, "2025-04-25T02:19:00-04:00", s.Value)
	})

	t.Run("marshal zero time", func(t *testing.T) {
		var ct CustomTime // Zero time
		av, err := ct.MarshalDynamoDBAttributeValue()
		require.NoError(t, err)

		s := av.(*types.AttributeValueMemberS)
		assert.Equal(t, time.Time{}.Format(time.RFC3339), s.Value)
	})

	t.Run("unmarshal valid string", func(t *testing.T) {
		av := &types.AttributeValueMemberS{
			Value: "2025-04-25T02:19:00-04:00",
		}

		var ct CustomTime
		err := ct.UnmarshalDynamoDBAttributeValue(av)
		require.NoError(t, err)
		expected := time.Date(2025, 4, 25, 2, 19, 0, 0, time.FixedZone("", -4*60*60))

		assert.True(t, expected.Equal(ct.Time), "Unmarshaled time should match")
	})

	t.Run("unmarshal invalid type", func(t *testing.T) {
		av := &types.AttributeValueMemberN{Value: "12345"}
		var ct CustomTime

		err := ct.UnmarshalDynamoDBAttributeValue(av)
		require.NoError(t, err)
		assert.True(t, ct.IsZero(), "Time should remain zero value")
	})

	t.Run("unmarshal invalid time string", func(t *testing.T) {
		av := &types.AttributeValueMemberS{
			Value: "invalid-time-format",
		}

		var ct CustomTime
		err := ct.UnmarshalDynamoDBAttributeValue(av)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot parse")
	})

	t.Run("roundtrip through attributevalue package", func(t *testing.T) {
		original := struct {
			EventTime CustomTime `dynamodbav:"eventTime"`
		}{
			EventTime: CustomTime{Time: time.Date(2025, 4, 25, 2, 19, 0, 0, time.UTC)},
		}

		item, err := attributevalue.MarshalMap(original)
		require.NoError(t, err)

		var result struct {
			EventTime CustomTime `dynamodbav:"eventTime"`
		}
		err = attributevalue.UnmarshalMap(item, &result)
		require.NoError(t, err)

		assert.Equal(t, original.EventTime.String(), result.EventTime.String())
	})
}
