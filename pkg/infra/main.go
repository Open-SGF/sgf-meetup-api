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

var EventsTableProps = DynamoDbProps{
	TableProps: &awsdynamodb.TableProps{
		TableName: jsii.String("MeetupEvents"),
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("Id"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	},
	GlobalSecondaryIndexes: []*awsdynamodb.GlobalSecondaryIndexProps{
		{
			IndexName: jsii.String("EventsByGroupIndex"),
			PartitionKey: &awsdynamodb.Attribute{
				Name: jsii.String("MeetupGroupUrlName"),
				Type: awsdynamodb.AttributeType_STRING,
			},
			SortKey: &awsdynamodb.Attribute{
				Name: jsii.String("EventDateTime"),
				Type: awsdynamodb.AttributeType_STRING,
			},
		},
	},
}

var ImportLogsTableProps = DynamoDbProps{
	TableProps: &awsdynamodb.TableProps{
		TableName: jsii.String("ImporterLog"),
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("Id"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		SortKey: &awsdynamodb.Attribute{
			Name: jsii.String("StartedAt"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	},
}

var MeetupFunctionName = jsii.String("meetupproxy")

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

	awsdynamodb.NewTable(stack, ImportLogsTableProps.TableName, ImportLogsTableProps.TableProps)

	customconstructs.NewGoLambdaFunction(stack, MeetupFunctionName, &customconstructs.GoLambdaFunctionProps{
		CodePath:     jsii.String("./cmd/meetupproxy"),
		FunctionName: MeetupFunctionName,
	})

	customconstructs.NewGoLambdaFunction(stack, jsii.String("importer"), &customconstructs.GoLambdaFunctionProps{
		CodePath: jsii.String("./cmd/importer"),
		Environment: &map[string]*string{
			"EVENTS_TABLE_NAME":          EventsTableProps.TableName,
			"IMPORTER_LOG_TABLE_NAME":    ImportLogsTableProps.TableName,
			"MEETUP_TOKEN_FUNCTION_NAME": MeetupFunctionName,
		},
	})

	customconstructs.NewGoLambdaFunction(stack, jsii.String("api"), &customconstructs.GoLambdaFunctionProps{
		CodePath: jsii.String("./cmd/api"),
	})

	return stack
}
