package infra

import (
	"fmt"
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3assets"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"os"
	"os/exec"
	"path/filepath"
)

type AppStackProps struct {
	awscdk.StackProps
}

const importerFunctionName = "importer"
const importerFunctionCodePath = "./cmd/importer/"

const defaultHandler = "main"
const defaultMemory = 128
const defaultTimeout = 60

func NewStack(scope constructs.Construct, id string, props *AppStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	awslambda.NewFunction(stack, jsii.String(importerFunctionName), &awslambda.FunctionProps{
		FunctionName: jsii.String(*stack.StackName() + "-" + importerFunctionName),
		Runtime:      awslambda.Runtime_GO_1_X(),
		MemorySize:   jsii.Number(defaultMemory),
		Timeout:      awscdk.Duration_Seconds(jsii.Number(defaultTimeout)),
		//Code: awslambda.AssetCode_FromAsset(jsii.String("."), &awss3assets.AssetOptions{
		//	Exclude: jsii.Strings("cdk.out", "cdk.json", ".nvmrc", ".gitignore", "package.json", "package-lock.json", "node_modules", ".idea"),
		//}),
		Code: awslambda.AssetCode_FromAsset(jsii.String("."), &awss3assets.AssetOptions{
			Bundling: &awscdk.BundlingOptions{
				Image: awscdk.DockerImage_FromBuild(jsii.String("./src/infra"), nil),
				Command: jsii.Strings(
					"bash", "-c",
					"go build -o /asset-output/main ./cmd/importer/main.go && "+
						"if [ -f ./cmd/importer/.env ]; then cp ./cmd/importer/.env /asset-output/; fi",
				),
				Volumes: &[]*awscdk.DockerVolume{
					{
						ContainerPath: jsii.String("/cache"),
						HostPath:      jsii.String("cache"),
					},
				},
			},
		}),
		Handler: jsii.String("main"),
	})

	return stack
}

type GoLocalBundling struct {
	FunctionPath string
	OutputName   string
}

func (g *GoLocalBundling) TryBundle(outputDir *string, options *awscdk.BundlingOptions) *bool {
	functionDir := filepath.Dir(g.FunctionPath)

	commands := []string{
		fmt.Sprintf("go build -o %s/%s %s", *outputDir, g.OutputName, g.FunctionPath),
		fmt.Sprintf("if [ -f %s/.env ]; then cp %s/.env %s/; fi", functionDir, functionDir, *outputDir),
	}

	for _, command := range commands {
		cmd := exec.Command("bash", "-c", command)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout

		if err := cmd.Run(); err != nil {
			success := false
			return &success
		}
	}

	success := true
	return &success
}
