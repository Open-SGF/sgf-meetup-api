package upsertuser

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"golang.org/x/crypto/bcrypt"
	"regexp"
	"sgf-meetup-api/pkg/shared/db"
	"sgf-meetup-api/pkg/shared/models"
	"sgf-meetup-api/pkg/shared/resource"
	"sgf-meetup-api/pkg/upsertuser/upsertuserconfig"
)

type Service struct {
	config *upsertuserconfig.Config
	db     *db.Client
}

func NewService(config *upsertuserconfig.Config, db *db.Client) *Service {
	return &Service{
		db:     db,
		config: config,
	}
}

func (s *Service) UpsertUser(ctx context.Context, tableName, clientID, clientSecret string) error {
	if err := s.validateClientSecret(clientSecret); err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(clientSecret), bcrypt.DefaultCost)

	if err != nil {
		return err
	}

	user := models.APIUser{
		ClientID:           clientID,
		HashedClientSecret: hash,
	}

	av, err := attributevalue.MarshalMap(user)

	if err != nil {
		return err
	}

	userTableNamer := resource.NewNamer(s.config.AppEnv, tableName)

	_, err = s.db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(userTableNamer.FullName()),
		Item:      av,
	})

	return err
}

const (
	minLength      = 12
	allowedSpecial = `!@#$%^&*()_+[]{};':"\|,.<>/?-`
)

var (
	hasUpper   = regexp.MustCompile(`[A-Z]`)
	hasLower   = regexp.MustCompile(`[a-z]`)
	hasNumber  = regexp.MustCompile(`[0-9]`)
	hasSpecial = regexp.MustCompile(`[` + regexp.QuoteMeta(allowedSpecial) + `]`)
	validChars = regexp.MustCompile(`^[A-Za-z0-9` + regexp.QuoteMeta(allowedSpecial) + `]+$`)
)

func (s *Service) validateClientSecret(clientSecret string) error {
	if len(clientSecret) < minLength {
		return fmt.Errorf("client secret must be at least %d characters", minLength)
	}

	if !hasUpper.MatchString(clientSecret) {
		return fmt.Errorf("client secret must contain at least one uppercase letter")
	}

	if !hasLower.MatchString(clientSecret) {
		return fmt.Errorf("client secret must contain at least one lowercase letter")
	}

	if !hasNumber.MatchString(clientSecret) {
		return fmt.Errorf("client secret must contain at least one number")
	}

	if !hasSpecial.MatchString(clientSecret) {
		return fmt.Errorf("client secret must contain at least one special character (%s)", allowedSpecial)
	}

	if !validChars.MatchString(clientSecret) {
		return fmt.Errorf("client secret contains invalid characters - only alphanumerics and %s are allowed", allowedSpecial)
	}
	return nil
}
