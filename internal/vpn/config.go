package vpn

import (
	"fmt"
	"github.com/pete911/aws-vpn/internal/aws/iam"
	"strings"
)

type Config struct {
	Name        string
	AccountId   string
	Region      string
	SecretsPath string
}

func GetSecretsPath(name string) string {
	return fmt.Sprintf("/vpn/%s/secrets", name)
}

func NewConfig(name, accountId, region string) Config {
	return Config{
		Name:        name,
		AccountId:   accountId,
		Region:      region,
		SecretsPath: GetSecretsPath(name),
	}
}

func (c Config) GetSecretsInlinePolicy() iam.InlinePolicyInput {
	resource := fmt.Sprintf("arn:aws:secretsmanager:%s:%s:secret:/%s/*",
		c.Region, c.AccountId, strings.Trim(c.SecretsPath, "/"))

	actions := []string{
		"secretsmanager:ListSecretVersionIds",
		"secretsmanager:GetSecretValue",
		"secretsmanager:DescribeSecret",
		"secretsmanager:CreateSecret",
		"secretsmanager:PutSecretValue",
		"secretsmanager:UpdateSecret",
	}
	return iam.NewInlinePolicyInput(c.Name, resource, actions)
}

func GetSSMManagedPolicies() []string {
	return []string{"AmazonSSMManagedInstanceCore", "AmazonSSMPatchAssociation"}
}
