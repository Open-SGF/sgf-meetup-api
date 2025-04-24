package infra

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscertificatemanager"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsevents"
	"github.com/aws/aws-cdk-go/awscdk/v2/awseventstargets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsssm"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"sgf-meetup-api/pkg/infra/customconstructs"
)

type AppStackProps struct {
	awscdk.StackProps
	AppEnv     string
	DomainName string
}

func NewStack(scope constructs.Construct, id string, props *AppStackProps) awscdk.Stack {
	namespace := props.AppEnv
	if namespace != "" {
		namespace = namespace + "-"
	}

	stackId := namespace + id
	stack := awscdk.NewStack(scope, &stackId, &props.StackProps)

	eventsTable := customconstructs.NewDynamoTable(stack, namespace, EventsTableProps)
	archivedEventsTable := customconstructs.NewDynamoTable(stack, namespace, ArchivedEventsTableProps)
	apiUsersTable := customconstructs.NewDynamoTable(stack, namespace, ApiUsersTableProps)

	meetupProxyFunctionName := customconstructs.NewFunctionName(namespace, "MeetupProxy")

	meetupPrivateKey := awsssm.StringParameter_ValueForStringParameter(stack, jsii.String("/sgf-meetup-api/meetup-private-key-base64"), nil)
	meetupUserId := awsssm.StringParameter_ValueForStringParameter(stack, jsii.String("/sgf-meetup-api/meetup-user-id"), nil)
	meetupClientId := awsssm.StringParameter_ValueForStringParameter(stack, jsii.String("/sgf-meetup-api/meetup-client-key"), nil)
	meetupSigningKeyId := awsssm.StringParameter_ValueForStringParameter(stack, jsii.String("/sgf-meetup-api/meetup-signing-key-id"), nil)

	meetupProxyFunction := customconstructs.NewGoLambdaFunction(stack, meetupProxyFunctionName.Name(), &customconstructs.GoLambdaFunctionProps{
		CodePath:     jsii.String("./cmd/meetupproxy"),
		FunctionName: meetupProxyFunctionName.PrefixedName(),
		Environment: &map[string]*string{
			"LOG_LEVEL":                 jsii.String("debug"),
			"LOG_TYPE":                  jsii.String("json"),
			"MEETUP_PRIVATE_KEY_BASE64": meetupPrivateKey,
			"MEETUP_USER_ID":            meetupUserId,
			"MEETUP_CLIENT_KEY":         meetupClientId,
			"MEETUP_SIGNING_KEY_ID":     meetupSigningKeyId,
		},
	})

	meetupProxyFunctionInvokePolicy := awsiam.NewManagedPolicy(stack, jsii.String("MeetupProxyFunctionInvokePolicy"), &awsiam.ManagedPolicyProps{
		ManagedPolicyName: jsii.String(namespace + "meetupProxyFunctionInvokePolicy"),
		Statements: &[]awsiam.PolicyStatement{awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
			Actions:   jsii.Strings("lambda:InvokeFunction"),
			Resources: jsii.Strings(*meetupProxyFunction.Function.FunctionArn()), //nolint:staticcheck
			Effect:    awsiam.Effect_ALLOW,
		})},
	})

	awscdk.NewCfnOutput(stack, jsii.String("MeetupProxyFunctionInvokePolicyArn"), &awscdk.CfnOutputProps{
		Value:       meetupProxyFunctionInvokePolicy.ManagedPolicyArn(),
		Description: jsii.String("ARN of the policy to invoke lambda"),
	})

	meetupGroupNames := awsssm.StringParameter_ValueForStringParameter(stack, jsii.String("/sgf-meetup-api/meetup-group-names"), nil)

	importerFunctionName := customconstructs.NewFunctionName(namespace, "Importer")

	importerFunction := customconstructs.NewGoLambdaFunction(stack, importerFunctionName.Name(), &customconstructs.GoLambdaFunctionProps{
		CodePath:     jsii.String("./cmd/importer"),
		FunctionName: importerFunctionName.PrefixedName(),
		Environment: &map[string]*string{
			"LOG_LEVEL":                     jsii.String("debug"),
			"LOG_TYPE":                      jsii.String("json"),
			"MEETUP_GROUP_NAMES":            meetupGroupNames,
			"MEETUP_PROXY_FUNCTION_NAME":    meetupProxyFunctionName.PrefixedName(),
			"EVENTS_TABLE_NAME":             &eventsTable.FullTableName,
			"GROUP_ID_DATE_TIME_INDEX_NAME": GroupIdDateTimeIndex.IndexName,
			"ARCHIVED_EVENTS_TABLE_NAME":    &archivedEventsTable.FullTableName,
		},
	})

	apiFunctionName := customconstructs.NewFunctionName(namespace, "Api")

	apiFunction := customconstructs.NewGoLambdaFunction(stack, apiFunctionName.Name(), &customconstructs.GoLambdaFunctionProps{
		CodePath:     jsii.String("./cmd/api"),
		FunctionName: apiFunctionName.PrefixedName(),
		Environment: &map[string]*string{
			"LOG_LEVEL":                     jsii.String("debug"),
			"LOG_TYPE":                      jsii.String("json"),
			"EVENTS_TABLE_NAME":             &eventsTable.FullTableName,
			"GROUP_ID_DATE_TIME_INDEX_NAME": GroupIdDateTimeIndex.IndexName,
			"API_USERS_TABLE_NAME":          &apiUsersTable.FullTableName,
		},
	})

	meetupProxyFunction.Function.GrantInvoke(importerFunction.Function)     //nolint:staticcheck
	eventsTable.Table.GrantReadWriteData(importerFunction.Function)         //nolint:staticcheck
	eventsTable.Table.GrantReadWriteData(apiFunction.Function)              //nolint:staticcheck
	archivedEventsTable.Table.GrantReadWriteData(importerFunction.Function) //nolint:staticcheck
	archivedEventsTable.Table.GrantReadWriteData(apiFunction.Function)      //nolint:staticcheck
	apiUsersTable.Table.GrantReadWriteData(apiFunction.Function)            //nolint:staticcheck

	importScheduleRule := awsevents.NewRule(stack, jsii.String("ImporterEventBridgeRule"), &awsevents.RuleProps{
		Schedule: awsevents.Schedule_Expression(jsii.String("cron(0 0-23/2 * * ? *)")), // every 2 hours
	})

	importScheduleRule.AddTarget(awseventstargets.NewLambdaFunction(
		importerFunction.Function,
		&awseventstargets.LambdaFunctionProps{},
	))

	api := awsapigateway.NewLambdaRestApi(stack, jsii.String("EventsGateway"), &awsapigateway.LambdaRestApiProps{
		Handler: apiFunction.Function,
		Proxy:   jsii.Bool(true),
	})

	certificate := awscertificatemanager.NewCertificate(stack, jsii.String("ApiCert"), &awscertificatemanager.CertificateProps{
		DomainName: jsii.String(props.DomainName),
		Validation: awscertificatemanager.CertificateValidation_FromDns(nil),
	})

	domain := awsapigateway.NewDomainName(stack, jsii.String("EventsGatewayDomain"), &awsapigateway.DomainNameProps{
		DomainName:   jsii.String(props.DomainName),
		Certificate:  certificate,
		EndpointType: awsapigateway.EndpointType_EDGE,
	})

	awsapigateway.NewBasePathMapping(stack, jsii.String("EventsGatewayBasePathMapping"), &awsapigateway.BasePathMappingProps{
		DomainName: domain,
		RestApi:    api,
		BasePath:   jsii.String(""),
	})

	awscdk.NewCfnOutput(stack, jsii.String("EventsGatewayDomainTarget"), &awscdk.CfnOutputProps{
		Value:       domain.DomainNameAliasDomainName(),
		Description: jsii.String("events gateway domain"),
	})

	return stack
}
