package infra

import (
	"fmt"
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
	"sgf-meetup-api/pkg/shared/resource"
)

type AppStackProps struct {
	awscdk.StackProps
	AppEnv     string
	DomainName string
}

func NewStack(scope constructs.Construct, id string, props *AppStackProps) awscdk.Stack {
	stackName := resource.NewNamer(props.AppEnv, id)

	stack := awscdk.NewStack(scope, jsii.String(stackName.FullName()), &props.StackProps)

	eventsTable := customconstructs.NewDynamoTable(stack, props.AppEnv, EventsTableProps)
	archivedEventsTable := customconstructs.NewDynamoTable(stack, props.AppEnv, ArchivedEventsTableProps)
	apiUsersTable := customconstructs.NewDynamoTable(stack, props.AppEnv, ApiUsersTableProps)

	commonEnvVars := map[string]*string{
		"LOG_LEVEL":                      jsii.String("debug"),
		"LOG_TYPE":                       jsii.String("json"),
		"DYNAMODB_ENDPOINT":              jsii.String(""),
		"DYNAMODB_AWS_REGION":            jsii.String(""),
		"DYNAMODB_AWS_ACCESS_KEY":        jsii.String(""),
		"DYNAMODB_AWS_SECRET_ACCESS_KEY": jsii.String(""),
	}

	meetupProxyFunctionName := resource.NewNamer(props.AppEnv, "MeetupProxy")

	meetupProxySSMPath := "/sgf-meetup-api/" + meetupProxyFunctionName.FullName()

	meetupProxyFunction := customconstructs.NewGoLambdaFunction(stack, jsii.String(meetupProxyFunctionName.Name()), &customconstructs.GoLambdaFunctionProps{
		CodePath:     jsii.String("./cmd/meetupproxy"),
		FunctionName: jsii.String(meetupProxyFunctionName.FullName()),
		Environment: &map[string]*string{
			"MEETUP_PRIVATE_KEY_BASE64": jsii.String(""),
			"MEETUP_USER_ID":            jsii.String(""),
			"MEETUP_CLIENT_KEY":         jsii.String(""),
			"MEETUP_SIGNING_KEY_ID":     jsii.String(""),
			"MEETUP_AUTH_URL":           jsii.String(""),
			"MEETUP_API_URL":            jsii.String(""),
			"SSM_PATH":                  jsii.String(meetupProxySSMPath),
		},
	})

	//nolint:staticcheck
	meetupProxyFunction.Function.AddToRolePolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect:    awsiam.Effect_ALLOW,
		Actions:   jsii.Strings("ssm:GetParameter", "ssm:GetParametersByPath"),
		Resources: jsii.Strings(fmt.Sprintf("arn:aws:ssm:%s:%s:parameter%s*", *awscdk.Aws_REGION(), *awscdk.Aws_ACCOUNT_ID(), meetupProxySSMPath)),
	}))

	meetupProxyFunctionPolicyNamer := resource.NewNamer(props.AppEnv, "MeetupProxyFunctionInvokePolicy")
	meetupProxyFunctionInvokePolicy := awsiam.NewManagedPolicy(stack, jsii.String(meetupProxyFunctionPolicyNamer.Name()), &awsiam.ManagedPolicyProps{
		ManagedPolicyName: jsii.String(meetupProxyFunctionPolicyNamer.FullName()),
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

	importerFunctionName := resource.NewNamer(props.AppEnv, "Importer")

	importerSSMPath := "/sgf-meetup-api/" + importerFunctionName.FullName()

	importerFunction := customconstructs.NewGoLambdaFunction(stack, jsii.String(importerFunctionName.Name()), &customconstructs.GoLambdaFunctionProps{
		CodePath:     jsii.String("./cmd/importer"),
		FunctionName: jsii.String(importerFunctionName.FullName()),
		Environment: mergeMaps(commonEnvVars, map[string]*string{
			"MEETUP_GROUP_NAMES":            meetupGroupNames,
			"MEETUP_PROXY_FUNCTION_NAME":    jsii.String(meetupProxyFunctionName.FullName()),
			"EVENTS_TABLE_NAME":             &eventsTable.FullTableName,
			"GROUP_ID_DATE_TIME_INDEX_NAME": GroupIdDateTimeIndex.IndexName,
			"ARCHIVED_EVENTS_TABLE_NAME":    &archivedEventsTable.FullTableName,
			"SSM_PATH":                      jsii.String(importerSSMPath),
		}),
	})

	//nolint:staticcheck
	importerFunction.Function.AddToRolePolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect:    awsiam.Effect_ALLOW,
		Actions:   jsii.Strings("ssm:GetParameter", "ssm:GetParametersByPath"),
		Resources: jsii.Strings(fmt.Sprintf("arn:aws:ssm:%s:%s:parameter%s*", *awscdk.Aws_REGION(), *awscdk.Aws_ACCOUNT_ID(), importerSSMPath)),
	}))

	apiFunctionName := resource.NewNamer(props.AppEnv, "Api")

	apiSSMPath := "/sgf-meetup-api/" + apiFunctionName.FullName()

	apiFunction := customconstructs.NewGoLambdaFunction(stack, jsii.String(apiFunctionName.Name()), &customconstructs.GoLambdaFunctionProps{
		CodePath:     jsii.String("./cmd/api"),
		FunctionName: jsii.String(apiFunctionName.FullName()),
		Environment: mergeMaps(commonEnvVars, map[string]*string{
			"EVENTS_TABLE_NAME":             &eventsTable.FullTableName,
			"GROUP_ID_DATE_TIME_INDEX_NAME": GroupIdDateTimeIndex.IndexName,
			"API_USERS_TABLE_NAME":          &apiUsersTable.FullTableName,
			"APP_URL":                       jsii.String("https://" + props.DomainName),
			"JWT_ISSUER":                    jsii.String(props.DomainName),
			"JWT_SECRET":                    jsii.String(""),
			"SSM_PATH":                      jsii.String(apiSSMPath),
		}),
	})

	//nolint:staticcheck
	apiFunction.Function.AddToRolePolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect:    awsiam.Effect_ALLOW,
		Actions:   jsii.Strings("ssm:GetParameter", "ssm:GetParametersByPath"),
		Resources: jsii.Strings(fmt.Sprintf("arn:aws:ssm:%s:%s:parameter%s*", *awscdk.Aws_REGION(), *awscdk.Aws_ACCOUNT_ID(), apiSSMPath)),
	}))

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
		Handler:       apiFunction.Function,
		Proxy:         jsii.Bool(true),
		EndpointTypes: &[]awsapigateway.EndpointType{awsapigateway.EndpointType_REGIONAL},
	})

	certificate := awscertificatemanager.NewCertificate(stack, jsii.String("ApiCert"), &awscertificatemanager.CertificateProps{
		DomainName: jsii.String(props.DomainName),
		Validation: awscertificatemanager.CertificateValidation_FromDns(nil),
	})

	domain := awsapigateway.NewDomainName(stack, jsii.String("EventsGatewayDomain"), &awsapigateway.DomainNameProps{
		DomainName:   jsii.String(props.DomainName),
		Certificate:  certificate,
		EndpointType: awsapigateway.EndpointType_REGIONAL,
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

func mergeMaps[M ~map[K]V, K comparable, V any](maps ...M) *M {
	merged := make(M)
	for _, m := range maps {
		for k, v := range m {
			merged[k] = v
		}
	}
	return &merged
}
