package configparser

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/spf13/viper"
	"strings"
)

const (
	SSMPathKey                  = "SSM_PATH"
	AWSRegionKey                = "AWS_REGION"
	AWSAccessKeyKey             = "AWS_ACCESS_KEY"
	AWSSecretAccessKeyKey       = "AWS_SECRET_ACCESS_KEY"
	AWSSessionTokenKey          = "AWS_SESSION_TOKEN"
	AWSProfileKey               = "AWS_PROFILE"
	AWSConfigFileKey            = "AWS_CONFIG_FILE"
	AWSSharedCredentialsFileKey = "AWS_SHARED_CREDENTIALS_FILE"
)

type configProcessor func(ctx context.Context, v *viper.Viper) error

type Parser struct {
	processors []configProcessor
}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) DefineKeys(keys []string) *Parser {
	p.processors = append(p.processors, func(ctx context.Context, v *viper.Viper) error {
		for _, key := range keys {
			v.SetDefault(strings.ToLower(key), "")
		}

		return nil
	})
	return p
}

func (p *Parser) WithEnvFile(path, filename string) *Parser {
	p.processors = append(p.processors, func(ctx context.Context, v *viper.Viper) error {
		v.SetConfigName(filename)
		v.SetConfigType("env")
		v.AddConfigPath(path)

		if err := v.ReadInConfig(); err != nil {
			var configFileNotFoundError viper.ConfigFileNotFoundError
			if !errors.As(err, &configFileNotFoundError) {
				return err
			}
		}

		return nil
	})

	return p
}

func (p *Parser) WithEnvVars() *Parser {
	p.processors = append(p.processors, func(ctx context.Context, v *viper.Viper) error {
		v.AutomaticEnv()

		return nil
	})

	return p
}

func (p *Parser) WithCustomProcessor(processor func(ctx context.Context, v *viper.Viper) error) *Parser {
	p.processors = append(p.processors, processor)
	return p
}

type SSMParameterOptions struct {
	Config  *aws.Config
	SSMPath string
}

func (p *Parser) WithSSMParameters(configure func(ctx context.Context, v *viper.Viper, opts *SSMParameterOptions)) *Parser {
	p.processors = append(p.processors, func(ctx context.Context, v *viper.Viper) error {
		ssmOptions := SSMParameterOptions{}
		configure(ctx, v, &ssmOptions)

		if ssmOptions.Config == nil || ssmOptions.SSMPath == "" {
			return nil
		}

		client := ssm.NewFromConfig(*ssmOptions.Config)

		paginator := ssm.NewGetParametersByPathPaginator(client, &ssm.GetParametersByPathInput{
			Path:           aws.String(ssmOptions.SSMPath),
			WithDecryption: aws.Bool(true),
		})

		parameters := make(map[string]string)

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				return err
			}

			for _, param := range page.Parameters {
				key := strings.TrimPrefix(*param.Name, ssmOptions.SSMPath)
				key = strings.ToLower(key)
				key = strings.TrimPrefix(key, "/")

				parameters[key] = *param.Value
			}
		}

		for key, value := range parameters {
			v.Set(key, value)
		}

		return nil
	})

	return p
}

func (p *Parser) Parse(ctx context.Context, output any) error {
	v := viper.New()

	for _, processor := range p.processors {
		if err := processor(ctx, v); err != nil {
			return err
		}
	}

	if err := v.Unmarshal(output); err != nil {
		return err
	}

	return nil
}
