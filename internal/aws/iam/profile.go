package iam

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
)

type InstanceProfileInput struct {
	Name string
	Tags map[string]string
	Role RoleInput
}

func (i InstanceProfileInput) toTags() []types.Tag {
	var out []types.Tag
	for k, v := range i.Tags {
		out = append(out, types.Tag{Key: aws.String(k), Value: aws.String(v)})
	}
	return out
}

type InstanceProfile struct {
	Arn                 string
	InstanceProfileId   string
	InstanceProfileName string
	Path                string
	RoleNames           []string
	Tags                map[string]string
}

func ToIamInstanceProfile(in *types.InstanceProfile) InstanceProfile {
	if in == nil {
		return InstanceProfile{}
	}
	var roleNames []string
	for _, role := range in.Roles {
		roleNames = append(roleNames, aws.ToString(role.RoleName))
	}

	return InstanceProfile{
		Arn:                 aws.ToString(in.Arn),
		InstanceProfileId:   aws.ToString(in.InstanceProfileId),
		InstanceProfileName: aws.ToString(in.InstanceProfileName),
		Path:                aws.ToString(in.Path),
		RoleNames:           roleNames,
		Tags:                fromTags(in.Tags),
	}
}
