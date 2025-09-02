package infra

import (
	"sgf-meetup-api/pkg/infra/customconstructs"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/jsii-runtime-go"
)

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

var EventsTableProps = &customconstructs.DynamoTableProps{
	TableProps: awsdynamodb.TableProps{
		TableName: jsii.String("MeetupEvents"),
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("id"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
		BillingMode:   awsdynamodb.BillingMode_PAY_PER_REQUEST,
	},
	GlobalSecondaryIndexes: []awsdynamodb.GlobalSecondaryIndexProps{
		GroupIdDateTimeIndex,
	},
}

var ArchivedEventsTableProps = &customconstructs.DynamoTableProps{
	TableProps: awsdynamodb.TableProps{
		TableName: jsii.String("MeetupArchivedEvents"),
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("id"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
		BillingMode:   awsdynamodb.BillingMode_PAY_PER_REQUEST,
	},
	GlobalSecondaryIndexes: []awsdynamodb.GlobalSecondaryIndexProps{
		GroupIdDateTimeIndex,
	},
}

var ApiUsersTableProps = &customconstructs.DynamoTableProps{
	TableProps: awsdynamodb.TableProps{
		TableName: jsii.String("MeetupApiUsers"),
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("clientId"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
		BillingMode:   awsdynamodb.BillingMode_PAY_PER_REQUEST,
	},
}

var Tables = []customconstructs.DynamoTableProps{
	*EventsTableProps,
	*ArchivedEventsTableProps,
	*ApiUsersTableProps,
}
