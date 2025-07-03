package iam

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/pete911/aws-vpn/internal/errs"
	"log/slog"
	"time"
)

type Service struct {
	logger *slog.Logger
	svc    *iam.Client
}

func NewService(logger *slog.Logger, cfg aws.Config) Service {
	return Service{
		logger: logger.With("component", "aws.iam.service"),
		svc:    iam.NewFromConfig(cfg),
	}
}

func (s Service) CreateInstanceProfile(ctx context.Context, in InstanceProfileInput) error {
	createProfileIn := &iam.CreateInstanceProfileInput{
		InstanceProfileName: aws.String(in.Name),
		Tags:                in.toTags(),
	}
	if _, err := s.svc.CreateInstanceProfile(ctx, createProfileIn); err != nil {
		return errs.FromAwsApi(err, "iam create-instance-profile")
	}
	s.logger.DebugContext(ctx, fmt.Sprintf("created %s instace profile", in.Name))

	if err := s.createEc2Role(ctx, in.Role); err != nil {
		return err
	}

	addRoleIn := &iam.AddRoleToInstanceProfileInput{
		InstanceProfileName: aws.String(in.Name),
		RoleName:            aws.String(in.Role.RoleName),
	}
	if _, err := s.svc.AddRoleToInstanceProfile(ctx, addRoleIn); err != nil {
		return errs.FromAwsApi(err, "iam add-role-to-instance-profile")
	}
	s.logger.DebugContext(ctx, fmt.Sprintf("added %s role to %s instance profile", in.Role.RoleName, in.Name))

	// insane shitty aws is able to list and describe instance profile, but run instance will report invalid name
	// or similar crap. we need to do classic old school sleep to get around AWS "eventual consistency" crap
	s.logger.DebugContext(ctx, "waiting 10 seconds for instance profile to become available")
	time.Sleep(10 * time.Second)
	return nil
}

func (s Service) DeleteInstanceProfile(ctx context.Context, name string) error {
	instanceProfileOut, err := s.svc.GetInstanceProfile(ctx, &iam.GetInstanceProfileInput{InstanceProfileName: aws.String(name)})
	if err != nil {
		return errs.FromAwsApi(err, "iam get-instance-profile")
	}
	instanceProfile := ToIamInstanceProfile(instanceProfileOut.InstanceProfile)

	for _, roleName := range instanceProfile.RoleNames {
		if _, err := s.svc.RemoveRoleFromInstanceProfile(ctx, &iam.RemoveRoleFromInstanceProfileInput{
			InstanceProfileName: aws.String(name),
			RoleName:            aws.String(roleName)},
		); err != nil {
			return errs.FromAwsApi(err, "iam remove-role-from-instance-profile")
		}
		s.logger.InfoContext(ctx, fmt.Sprintf("removed %s role from %s instance profile", roleName, name))
		if err := s.deleteRole(ctx, roleName); err != nil {
			return err
		}
	}
	if _, err := s.svc.DeleteInstanceProfile(ctx, &iam.DeleteInstanceProfileInput{InstanceProfileName: aws.String(name)}); err != nil {
		return errs.FromAwsApi(err, "iam delete-instance-profile")
	}
	s.logger.InfoContext(ctx, fmt.Sprintf("deleted %s instance profile", name))
	return nil
}

func (s Service) deleteRole(ctx context.Context, name string) error {
	inlinePolicies, err := s.svc.ListRolePolicies(ctx, &iam.ListRolePoliciesInput{RoleName: aws.String(name)})
	if err != nil {
		return errs.FromAwsApi(err, "iam list-role-policies")
	}
	for _, policy := range inlinePolicies.PolicyNames {
		if _, err := s.svc.DeleteRolePolicy(ctx, &iam.DeleteRolePolicyInput{
			RoleName:   aws.String(name),
			PolicyName: aws.String(policy),
		}); err != nil {
			return errs.FromAwsApi(err, "iam delete-role-policy")
		}
		s.logger.InfoContext(ctx, fmt.Sprintf("deleted %s policy from %s role", policy, name))
	}

	managedPolicies, err := s.svc.ListAttachedRolePolicies(ctx, &iam.ListAttachedRolePoliciesInput{RoleName: aws.String(name)})
	if err != nil {
		return errs.FromAwsApi(err, "iam list-attached-role-policies")
	}
	for _, policy := range managedPolicies.AttachedPolicies {
		if _, err := s.svc.DetachRolePolicy(ctx, &iam.DetachRolePolicyInput{
			RoleName:  aws.String(name),
			PolicyArn: policy.PolicyArn,
		}); err != nil {
			return errs.FromAwsApi(err, "iam detach-role-policy")
		}
		s.logger.InfoContext(ctx, fmt.Sprintf("detached %s policy from %s role", aws.ToString(policy.PolicyName), name))
	}

	if _, err := s.svc.DeleteRole(ctx, &iam.DeleteRoleInput{RoleName: aws.String(name)}); err != nil {
		return errs.FromAwsApi(err, "iam delete-role")
	}
	s.logger.InfoContext(ctx, fmt.Sprintf("deleted %s role", name))
	return nil
}

func (s Service) createEc2Role(ctx context.Context, in RoleInput) error {
	roleIn := &iam.CreateRoleInput{
		AssumeRolePolicyDocument: aws.String(ec2AssumeRolePolicyDocument),
		RoleName:                 aws.String(in.RoleName),
		Description:              aws.String("aws-vpn role"),
		Tags:                     in.toTags(),
	}
	if _, err := s.svc.CreateRole(ctx, roleIn); err != nil {
		return errs.FromAwsApi(err, "iam create-role")
	}
	s.logger.DebugContext(ctx, fmt.Sprintf("created %s role", in.RoleName))

	if err := s.attachRolePolicy(ctx, in.RoleName, in.ManagedPolicyNames); err != nil {
		return err
	}
	if err := s.putRolePolicies(ctx, in.RoleName, in.InlinePolicies); err != nil {
		return err
	}
	return nil
}

func (s Service) attachRolePolicy(ctx context.Context, roleName string, policyNames []string) error {
	if len(policyNames) == 0 {
		s.logger.DebugContext(ctx, fmt.Sprintf("no managed policies provided, skipping attach role policy to %s role", roleName))
		return nil
	}

	for _, policyName := range policyNames {
		policyArn := fmt.Sprintf("arn:aws:iam::aws:policy/%s", policyName)
		in := &iam.AttachRolePolicyInput{RoleName: aws.String(roleName), PolicyArn: aws.String(policyArn)}
		if _, err := s.svc.AttachRolePolicy(ctx, in); err != nil {
			return errs.FromAwsApi(err, "iam attach-role-policy")
		}
		s.logger.DebugContext(ctx, fmt.Sprintf("attached %s role policy to %s role", policyArn, roleName))
	}
	return nil
}

func (s Service) putRolePolicies(ctx context.Context, roleName string, policies []InlinePolicyInput) error {
	if len(policies) == 0 {
		s.logger.DebugContext(ctx, fmt.Sprintf("no inline policies provided, skipping put role policy to %s role", roleName))
		return nil
	}

	for _, policy := range policies {
		in := &iam.PutRolePolicyInput{
			PolicyDocument: aws.String(policy.Document),
			PolicyName:     aws.String(policy.Name),
			RoleName:       aws.String(roleName),
		}
		if _, err := s.svc.PutRolePolicy(ctx, in); err != nil {
			return errs.FromAwsApi(err, "iam put-role-policy")
		}
		s.logger.DebugContext(ctx, fmt.Sprintf("put %s policy to %s role", policy.Name, roleName))
	}
	return nil
}
