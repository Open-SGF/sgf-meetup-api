package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/jsii-runtime-go"
	"log"
	"sgf-meetup-api/pkg/infra"
	"sgf-meetup-api/pkg/infra/infraconfig"
)

func main() {
	defer jsii.Close()

	config, err := infraconfig.NewConfig()

	if err != nil {
		log.Println(err)
	}

	app := awscdk.NewApp(nil)

	infra.NewStack(app, "SgfMeetupApiGo", &infra.AppStackProps{
		StackProps: awscdk.StackProps{
			Env: env(),
		},
		AppEnv: config.AppEnv,
	})

	app.Synth(nil)
}

func env() *awscdk.Environment {
	return nil
}
