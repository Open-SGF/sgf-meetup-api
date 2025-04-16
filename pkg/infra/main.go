package infra

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"sgf-meetup-api/pkg/infra/customconstructs"
	"sgf-meetup-api/pkg/shared/db"
)

type AppStackProps struct {
	awscdk.StackProps
}

var MeetupProxyFunctionName = jsii.String("meetupproxy")

func NewStack(scope constructs.Construct, id string, props *AppStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	eventsTable := awsdynamodb.NewTable(stack, db.EventsTableProps.TableName, db.EventsTableProps.TableProps)

	for _, gsi := range db.EventsTableProps.GlobalSecondaryIndexes {
		eventsTable.AddGlobalSecondaryIndex(gsi)
	}

	archivedEventsTable := awsdynamodb.NewTable(stack, db.ArchivedEventsTableProps.TableName, db.ArchivedEventsTableProps.TableProps)

	for _, gsi := range db.ArchivedEventsTableProps.GlobalSecondaryIndexes {
		archivedEventsTable.AddGlobalSecondaryIndex(gsi)
	}

	apiUsers := awsdynamodb.NewTable(stack, db.ApiUsersTableProps.TableName, db.ApiUsersTableProps.TableProps)

	for _, gsi := range db.ApiUsersTableProps.GlobalSecondaryIndexes {
		apiUsers.AddGlobalSecondaryIndex(gsi)
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
