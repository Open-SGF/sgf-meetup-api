package syncdynamodb

import (
	"github.com/spf13/viper"
	"sgf-meetup-api/pkg/configparser"
	"strings"
)

const (
	dynamoDbEndpoint   = "DYNAMODB_ENDPOINT"
	awsRegion          = "AWS_REGION"
	awsAccessKey       = "AWS_ACCESS_KEY"
	awsSecretAccessKey = "AWS_SECRET_ACCESS_KEY"
)

var keys = []string{
	strings.ToLower(dynamoDbEndpoint),
	strings.ToLower(awsRegion),
	strings.ToLower(awsAccessKey),
	strings.ToLower(awsSecretAccessKey),
}

type Config struct {
	DynamoDbEndpoint   string `mapstructure:"dynamodb_endpoint"`
	AwsRegion          string `mapstructure:"aws_region"`
	AwsAccessKey       string `mapstructure:"aws_access_key"`
	AwsSecretAccessKey string `mapstructure:"aws_secret_access_key"`
}

func NewConfig() (*Config, error) {
	return NewConfigFromEnvFile(".", ".env")
}

func NewConfigFromEnvFile(path, filename string) (*Config, error) {
	config, err := configparser.Parse[Config](configparser.ParseOptions{
		EnvFilepath: path,
		EnvFilename: filename,
		Keys:        keys,
		SetDefaults: setDefaults,
	})

	if err != nil {
		return nil, err
	}

	return config, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault(strings.ToLower(awsRegion), "us-east-2")
}
