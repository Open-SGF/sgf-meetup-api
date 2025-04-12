package infra

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"sgf-meetup-api/pkg/infra/customconstructs"
)

type AppStackProps struct {
	awscdk.StackProps
}

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

var MeetupProxyFunctionName = jsii.String("meetupproxy")

func NewStack(scope constructs.Construct, id string, props *AppStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	eventsTable := awsdynamodb.NewTable(stack, EventsTableProps.TableName, EventsTableProps.TableProps)

	for _, gsi := range EventsTableProps.GlobalSecondaryIndexes {
		eventsTable.AddGlobalSecondaryIndex(gsi)
	}

	archivedEventsTable := awsdynamodb.NewTable(stack, ArchivedEventsTableProps.TableName, ArchivedEventsTableProps.TableProps)

	for _, gsi := range ArchivedEventsTableProps.GlobalSecondaryIndexes {
		archivedEventsTable.AddGlobalSecondaryIndex(gsi)
	}

	customconstructs.NewGoLambdaFunction(stack, MeetupProxyFunctionName, &customconstructs.GoLambdaFunctionProps{
		CodePath:     jsii.String("./cmd/meetupproxy"),
		FunctionName: MeetupProxyFunctionName,
	})

	customconstructs.NewGoLambdaFunction(stack, jsii.String("importer"), &customconstructs.GoLambdaFunctionProps{
		CodePath: jsii.String("./cmd/importer"),
	})

	customconstructs.NewGoLambdaFunction(stack, jsii.String("api"), &customconstructs.GoLambdaFunctionProps{
		CodePath: jsii.String("./cmd/api"),
	})

	return stack
}
