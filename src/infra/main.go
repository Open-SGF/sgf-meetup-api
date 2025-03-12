package infra

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"sgf-meetup-api/src/infra/customconstructs"
)

type AppStackProps struct {
	awscdk.StackProps
}

func NewStack(scope constructs.Construct, id string, props *AppStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	eventsTable := awsdynamodb.NewTable(stack, jsii.String("MeetupEvents"), &awsdynamodb.TableProps{
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("Id"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		TableName:     jsii.String("MeetupEvents"),
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})

	eventsTableIndex := &awsdynamodb.GlobalSecondaryIndexProps{
		IndexName: jsii.String("EventsByGroupIndex"),
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("MeetupGroupUrlName"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		SortKey: &awsdynamodb.Attribute{
			Name: jsii.String("EventDateTime"),
			Type: awsdynamodb.AttributeType_STRING,
		},
	}

	eventsTable.AddGlobalSecondaryIndex(eventsTableIndex)

	importerLogsTable := awsdynamodb.NewTable(stack, jsii.String("ImporterLog"), &awsdynamodb.TableProps{
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("Id"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		SortKey: &awsdynamodb.Attribute{
			Name: jsii.String("StartedAt"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		TableName:     jsii.String("ImporterLog"),
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})

	customconstructs.NewGoLambdaFunction(stack, jsii.String("importer"), &customconstructs.GoLambdaFunctionProps{
		CodePath: jsii.String("./cmd/importer"),
		Environment: &map[string]*string{
			"EVENTS_TABLE_NAME":       eventsTable.TableName(),
			"IMPORTER_LOG_TABLE_NAME": importerLogsTable.TableName(),
		},
	})

	customconstructs.NewGoLambdaFunction(stack, jsii.String("meetuptoken"), &customconstructs.GoLambdaFunctionProps{
		CodePath: jsii.String("./cmd/meetuptoken"),
	})

	customconstructs.NewGoLambdaFunction(stack, jsii.String("api"), &customconstructs.GoLambdaFunctionProps{
		CodePath: jsii.String("./cmd/api"),
	})

	return stack
}
