package secrets

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/pete911/aws-vpn/internal/errs"
	"log/slog"
	"strings"
)

type Service struct {
	logger *slog.Logger
	svc    *secretsmanager.Client
}

func NewService(logger *slog.Logger, cfg aws.Config) Service {
	return Service{
		logger: logger.With("component", "aws.vpc.service"),
		svc:    secretsmanager.NewFromConfig(cfg),
	}
}

func (s Service) DeleteSecrets(ctx context.Context, prefix string) error {
	secrets, err := s.listSecrets(ctx, prefix)
	if err != nil {
		return err
	}
	for _, secret := range secrets {
		if _, err := s.svc.DeleteSecret(ctx, &secretsmanager.DeleteSecretInput{
			SecretId:                   aws.String(secret),
			ForceDeleteWithoutRecovery: aws.Bool(true),
		}); err != nil {
			return errs.FromAwsApi(err, "secretsmanager delete-secret")
		}
		s.logger.InfoContext(ctx, fmt.Sprintf("deleted %s secret", secret))
	}
	return nil
}

func (s Service) GetSecrets(ctx context.Context, prefix string) (map[string]string, error) {
	secretNames, err := s.listSecrets(ctx, prefix)
	if err != nil {
		return nil, err
	}

	secrets := make(map[string]string)
	for _, secretName := range secretNames {
		out, err := s.svc.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{SecretId: aws.String(secretName)})
		if err != nil {
			return nil, errs.FromAwsApi(err, "secretsmanager get-secret-value")
		}
		secrets[strings.TrimPrefix(secretName, prefix)] = aws.ToString(out.SecretString)
	}
	return secrets, nil
}

func (s Service) listSecrets(ctx context.Context, prefix string) ([]string, error) {
	var secrets []string
	in := &secretsmanager.ListSecretsInput{}
	for {
		out, err := s.svc.ListSecrets(ctx, in)
		if err != nil {
			return nil, errs.FromAwsApi(err, "secretsmanager list-secrets")
		}
		for _, secret := range out.SecretList {
			secretName := aws.ToString(secret.Name)
			if strings.HasPrefix(secretName, prefix) {
				secrets = append(secrets, secretName)
			}
		}
		if aws.ToString(out.NextToken) == "" {
			break
		}
		in.NextToken = out.NextToken
	}
	return secrets, nil
}
