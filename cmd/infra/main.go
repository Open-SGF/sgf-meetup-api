package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/jsii-runtime-go"
	"sgf-meetup-api/pkg/infra"
)

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	infra.NewStack(app, "SgfMeetupApiGo", &infra.AppStackProps{
		StackProps: awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

func env() *awscdk.Environment {
	return nil
}
