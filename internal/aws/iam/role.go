package iam

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
)

type RoleInput struct {
	RoleName           string
	ManagedPolicyNames []string
	InlinePolicies     []InlinePolicyInput
	Tags               map[string]string
}

func (r RoleInput) toTags() []types.Tag {
	var out []types.Tag
	for k, v := range r.Tags {
		out = append(out, types.Tag{Key: aws.String(k), Value: aws.String(v)})
	}
	return out
}
