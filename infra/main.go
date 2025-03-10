package infra

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/constructs-go/constructs/v10"
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

	// The code that defines your stack goes here

	// example resource
	// queue := awssqs.NewQueue(stack, jsii.String("InfraQueue"), &awssqs.QueueProps{
	// 	VisibilityTimeout: awscdk.Duration_Seconds(jsii.Number(300)),
	// })

	return stack
}
