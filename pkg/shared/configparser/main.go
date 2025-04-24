package configparser

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
)

type ParseOptions struct {
	EnvFilename string
	EnvFilepath string
	Keys        []string
	SetDefaults func(v *viper.Viper) error
	SSMPath     string
}

func Parse[T any](ctx context.Context, options ParseOptions) (*T, error) {
	v := viper.New()

	for _, key := range options.Keys {
		v.SetDefault(strings.ToLower(key), "")
	}

	v.SetConfigName(options.EnvFilename)
	v.SetConfigType("env")
	v.AddConfigPath(options.EnvFilepath)

	if err := v.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, err
		}
	}

	if isLambda() && options.SSMPath != "" {
		params, err := getSSMParameters(ctx, options.SSMPath)
		if err != nil {
			return nil, err
		}
		for key, value := range params {
			v.Set(key, value)
		}
	}

	v.AutomaticEnv()

	if err := options.SetDefaults(v); err != nil {
		return nil, err
	}

	var cfg T
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func ParseFromKey[T any](v *viper.Viper, key string, parser func(string) (T, error), fallback T) {
	normalizedKey := strings.ToLower(key)
	str := v.GetString(normalizedKey)
	value, err := parser(str)

	if err != nil {
		v.Set(normalizedKey, fallback)
		return
	}

	v.Set(normalizedKey, value)
}

func SetupTestEnv(envContent string) (string, func(), error) {
	tempDir, err := os.MkdirTemp("", "configtest")

	if err != nil {
		return "", nil, err
	}

	envPath := filepath.Join(tempDir, ".env")
	err = os.WriteFile(envPath, []byte(envContent), 0644)

	if err != nil {
		return "", nil, err
	}

	return tempDir, func() {
		_ = os.RemoveAll(tempDir)
	}, nil
}

func isLambda() bool {
	return os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != ""
}

func getSSMParameters(ctx context.Context, path string) (map[string]string, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic(err)
	}

	client := ssm.NewFromConfig(cfg)

	paginator := ssm.NewGetParametersByPathPaginator(client, &ssm.GetParametersByPathInput{
		Path:           aws.String(path),
		WithDecryption: aws.Bool(true),
	})

	parameters := make(map[string]string)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get SSM parameters page: %w", err)
		}

		for _, param := range page.Parameters {
			key := strings.TrimPrefix(*param.Name, path)
			key = strings.ToLower(key)
			key = strings.TrimPrefix(key, "/")

			parameters[key] = *param.Value
		}
	}

	return parameters, nil
}
