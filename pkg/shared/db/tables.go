package db

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/jsii-runtime-go"
)

type DynamoDbProps struct {
	*awsdynamodb.TableProps
	GlobalSecondaryIndexes []*awsdynamodb.GlobalSecondaryIndexProps
}

var GroupIdDateTimeIndex = awsdynamodb.GlobalSecondaryIndexProps{
	IndexName: jsii.String("GroupIdDateTimeIndex"),
	PartitionKey: &awsdynamodb.Attribute{
		Name: jsii.String("groupId"),
		Type: awsdynamodb.AttributeType_STRING,
	},
	SortKey: &awsdynamodb.Attribute{
		Name: jsii.String("dateTime"),
		Type: awsdynamodb.AttributeType_STRING,
	},
}

var EventsTableProps = DynamoDbProps{
	TableProps: &awsdynamodb.TableProps{
		TableName: jsii.String("MeetupEvents"),
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("id"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	},
	GlobalSecondaryIndexes: []*awsdynamodb.GlobalSecondaryIndexProps{
		&GroupIdDateTimeIndex,
	},
}

var ArchivedEventsTableProps = DynamoDbProps{
	TableProps: &awsdynamodb.TableProps{
		TableName: jsii.String("MeetupArchivedEvents"),
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("id"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	},
	GlobalSecondaryIndexes: []*awsdynamodb.GlobalSecondaryIndexProps{
		&GroupIdDateTimeIndex,
	},
}

var ApiUsersTableProps = DynamoDbProps{
	TableProps: &awsdynamodb.TableProps{
		TableName: jsii.String("MeeupApiUsers"),
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("id"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	},
}

var Tables = []DynamoDbProps{EventsTableProps, ArchivedEventsTableProps, ApiUsersTableProps}
