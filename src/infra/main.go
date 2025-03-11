package infra

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"sgf-meetup-api/src/infra/custom_constructs"
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

	custom_constructs.NewGoLambdaFunction(stack, "importer", &custom_constructs.GoLambdaFunctionProps{
		CodePath: jsii.String("./cmd/importer"),
	})

	custom_constructs.NewGoLambdaFunction(stack, "get_token", &custom_constructs.GoLambdaFunctionProps{
		CodePath: jsii.String("./cmd/get_token"),
	})

	custom_constructs.NewGoLambdaFunction(stack, "api", &custom_constructs.GoLambdaFunctionProps{
		CodePath: jsii.String("./cmd/api"),
	})

	return stack
}
