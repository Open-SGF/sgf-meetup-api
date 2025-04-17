package infra

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"sgf-meetup-api/pkg/infra/customconstructs"
)

type AppStackProps struct {
	awscdk.StackProps
	AppEnv string
}

func NewStack(scope constructs.Construct, id string, props *AppStackProps) awscdk.Stack {
	stack := awscdk.NewStack(scope, &id, &props.StackProps)

	namespace := props.AppEnv
	if namespace != "" {
		namespace = namespace + "-"
	}

	eventsTable := customconstructs.NewDynamoTable(stack, namespace, EventsTableProps)
	archivedEventsTable := customconstructs.NewDynamoTable(stack, namespace, ArchivedEventsTableProps)
	apiUsersTable := customconstructs.NewDynamoTable(stack, namespace, ApiUsersTableProps)

	meetupProxyFunctionName := customconstructs.NewFunctionName(namespace, "meetupproxy")

	customconstructs.NewGoLambdaFunction(stack, meetupProxyFunctionName.Name(), &customconstructs.GoLambdaFunctionProps{
		CodePath:     jsii.String("./cmd/meetupproxy"),
		FunctionName: meetupProxyFunctionName.PrefixedName(),
	})

	importerFunctionName := customconstructs.NewFunctionName(namespace, "importer")

	customconstructs.NewGoLambdaFunction(stack, importerFunctionName.Name(), &customconstructs.GoLambdaFunctionProps{
		CodePath:     jsii.String("./cmd/importer"),
		FunctionName: importerFunctionName.PrefixedName(),
		Environment: &map[string]*string{
			"MEETUP_PROXY_FUNCTION_NAME":    meetupProxyFunctionName.PrefixedName(),
			"EVENTS_TABLE_NAME":             &eventsTable.FullTableName,
			"GROUP_ID_DATE_TIME_INDEX_NAME": GroupIdDateTimeIndex.IndexName,
			"ARCHIVED_EVENTS_TABLE_NAME":    &archivedEventsTable.FullTableName,
		},
	})

	apiFunctionName := customconstructs.NewFunctionName(namespace, "api")

	customconstructs.NewGoLambdaFunction(stack, apiFunctionName.PrefixedName(), &customconstructs.GoLambdaFunctionProps{
		CodePath:     jsii.String("./cmd/api"),
		FunctionName: apiFunctionName.PrefixedName(),
		Environment: &map[string]*string{
			"EVENTS_TABLE_NAME":             &eventsTable.FullTableName,
			"GROUP_ID_DATE_TIME_INDEX_NAME": GroupIdDateTimeIndex.IndexName,
			"API_USERS_TABLE_NAME":          &apiUsersTable.FullTableName,
		},
	})

	return stack
}
