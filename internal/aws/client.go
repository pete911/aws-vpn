package aws

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/pete911/aws-vpn/internal/aws/iam"
	"github.com/pete911/aws-vpn/internal/aws/secrets"
	"github.com/pete911/aws-vpn/internal/aws/vpc"
	"github.com/pete911/aws-vpn/internal/errs"
	"log/slog"
	"strings"
	"time"
)

const ssmImageId = "resolve:ssm:/aws/service/ami-amazon-linux-latest/al2023-ami-kernel-default-x86_64"

type Client struct {
	AccountId  string
	Region     string
	logger     *slog.Logger
	vpcSvc     vpc.Service
	iamSvc     iam.Service
	secretsSvc secrets.Service
	ec2Svc     *ec2.Client
}

func NewClient(logger *slog.Logger, region string) (Client, error) {
	cfg, err := newAWSConfig("")
	if err != nil {
		return Client{}, err
	}
	if region != "" {
		cfg.Region = region
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	out, err := sts.NewFromConfig(cfg).GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return Client{}, errs.FromAwsApi(err, "sts get-caller-identity")
	}

	return Client{
		logger:     logger.With("component", "aws.client"),
		AccountId:  aws.ToString(out.Account),
		Region:     region,
		vpcSvc:     vpc.NewService(logger, cfg),
		iamSvc:     iam.NewService(logger, cfg),
		secretsSvc: secrets.NewService(logger, cfg),
		ec2Svc:     ec2.NewFromConfig(cfg),
	}, nil
}

func (c Client) GetDefaultPublicSubnets(ctx context.Context) (vpc.Subnets, error) {
	return c.vpcSvc.GetDefaultPublicSubnets(ctx)
}

func (c Client) GetSecrets(ctx context.Context, prefix string) (map[string]string, error) {
	return c.secretsSvc.GetSecrets(ctx, prefix)
}

func (c Client) TerminateInstance(ctx context.Context, in Instance, secretsPath string) error {
	// get instance that matches project tags and the name
	instance, err := c.DescribeInstanceById(ctx, in.Id)
	if err != nil {
		return err
	}

	if _, err := c.ec2Svc.TerminateInstances(ctx, &ec2.TerminateInstancesInput{InstanceIds: []string{instance.Id}}); err != nil {
		return errs.FromAwsApi(err, "ec2 terminate-instance")
	}
	c.logger.InfoContext(ctx, fmt.Sprintf("terminating instace %s", instance.Id))

	if err := c.iamSvc.DeleteInstanceProfile(ctx, instance.InstanceProfile); err != nil {
		return err
	}
	if err := c.secretsSvc.DeleteSecrets(ctx, secretsPath); err != nil {
		return err
	}

	// wait for instance to terminate
	for x := 0; x < 10; x++ {
		time.Sleep(10 * time.Second)
		status, err := c.DescribeInstanceStatus(ctx, instance.Id)
		if err != nil {
			return err
		}
		c.logger.InfoContext(ctx, fmt.Sprintf("instance %s state %s", instance.Id, status.InstanceState))
		if status.InstanceState == "terminated" {
			break
		}
	}

	// sometimes it takes longer for ENI to disappear, retry deleting of security group
	// if this happens, add retry loop
	for _, sg := range instance.SecurityGroups {
		// has to use security group id, name only works in default VPC
		if _, err := c.ec2Svc.DeleteSecurityGroup(ctx, &ec2.DeleteSecurityGroupInput{GroupId: aws.String(sg.Id)}); err != nil {
			return errs.FromAwsApi(err, "ec2 delete-security-group")
		}
		c.logger.InfoContext(ctx, fmt.Sprintf("deleted %s %s security group", sg.Name, sg.Id))
	}
	return nil
}

func (c Client) RunInstance(ctx context.Context, v RunInstancesInput) (Instance, error) {
	// first check if there is instance with the same tags
	filters := append(v.Metadata.toTagFilter(), types.Filter{Name: aws.String("instance-state-name"), Values: []string{"running"}})
	instances, err := c.describeInstances(ctx, filters)
	if err != nil {
		return Instance{}, err
	}
	if len(instances) != 0 {
		return Instance{}, fmt.Errorf("instance with %s name already exists", v.Metadata.Name)
	}

	securityGroupId, err := c.createSecurityGroup(ctx, v.Metadata, v.InboundCidr, v.InboundPort)
	if err != nil {
		return Instance{}, err
	}

	if err := c.iamSvc.CreateInstanceProfile(ctx, v.InstanceProfile); err != nil {
		return Instance{}, err
	}

	in := &ec2.RunInstancesInput{
		MaxCount: aws.Int32(1),
		MinCount: aws.Int32(1),
		IamInstanceProfile: &types.IamInstanceProfileSpecification{
			Name: aws.String(v.Metadata.Name),
		},
		ImageId:          aws.String(ssmImageId),
		InstanceType:     types.InstanceTypeT3Micro,
		SecurityGroupIds: []string{securityGroupId},
		SubnetId:         aws.String(v.SubnetId),
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceTypeInstance,
				Tags:         v.Metadata.toTags(),
			},
		},
		UserData: aws.String(base64.StdEncoding.EncodeToString([]byte(v.UserData))),
	}

	out, err := c.ec2Svc.RunInstances(ctx, in)
	if err != nil {
		return Instance{}, errs.FromAwsApi(err, "ec2 run-instances")
	}
	if len(out.Instances) != 1 {
		return Instance{}, fmt.Errorf("expected 1 instance, got %d", len(out.Instances))
	}

	instance := ToInstance(out.Instances[0])
	c.logger.DebugContext(ctx, fmt.Sprintf("launching instace %s", instance.Id))
	return instance, nil
}

func (c Client) DescribeInstanceStatus(ctx context.Context, id string) (InstanceStatus, error) {
	in := &ec2.DescribeInstanceStatusInput{InstanceIds: []string{id}, IncludeAllInstances: aws.Bool(true)}
	out, err := c.ec2Svc.DescribeInstanceStatus(ctx, in)
	if err != nil {
		return InstanceStatus{}, errs.FromAwsApi(err, "ec2 describe-instance-status")
	}
	if len(out.InstanceStatuses) != 1 {
		return InstanceStatus{}, fmt.Errorf("expected 1 instance status, got %d", len(out.InstanceStatuses))
	}
	return ToInstanceStatus(out.InstanceStatuses[0]), nil
}

func (c Client) DescribeInstanceById(ctx context.Context, id string) (Instance, error) {
	filters := []types.Filter{{Name: aws.String("instance-id"), Values: []string{id}}}
	instances, err := c.describeInstances(ctx, filters)
	if err != nil {
		return Instance{}, err
	}
	if len(instances) != 1 {
		return Instance{}, fmt.Errorf("expected 1 instance, got %d", len(instances))
	}
	return instances[0], nil
}

func (c Client) DescribeInstancesByNamePrefix(ctx context.Context, prefix string, tags map[string]string) (Instances, error) {
	if _, ok := tags["Name"]; ok {
		delete(tags, "Name")
	}
	filters := []types.Filter{{Name: aws.String("instance-state-name"), Values: []string{"running"}}}
	for k, v := range tags {
		filters = append(filters, types.Filter{Name: aws.String(fmt.Sprintf("tag:%s", k)), Values: []string{v}})
	}

	instances, err := c.describeInstances(ctx, filters)
	if err != nil {
		return nil, err
	}
	var filteredInstances Instances
	for _, instance := range instances {
		if strings.HasPrefix(instance.Name, prefix) {
			filteredInstances = append(filteredInstances, instance)
		}
	}
	return filteredInstances, nil
}

func (c Client) describeInstances(ctx context.Context, filters []types.Filter) (Instances, error) {
	in := &ec2.DescribeInstancesInput{Filters: filters}
	var instances Instances
	for {
		out, err := c.ec2Svc.DescribeInstances(ctx, in)
		if err != nil {
			return nil, errs.FromAwsApi(err, "ec2 describe-instances")
		}
		for _, reservation := range out.Reservations {
			instances = append(instances, ToInstances(reservation.Instances)...)
		}
		if aws.ToString(out.NextToken) == "" {
			break
		}
		in.NextToken = out.NextToken
	}
	c.logger.DebugContext(ctx, fmt.Sprintf("described %d instaces", len(instances)))
	return instances, nil
}

func (c Client) createSecurityGroup(ctx context.Context, in MetadataInput, inboundCidr string, inboundPort int) (string, error) {
	sgIn := &ec2.CreateSecurityGroupInput{
		GroupName:   aws.String(in.Name),
		Description: aws.String("aws-vpn project"),
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceTypeSecurityGroup,
				Tags:         in.toTags(),
			},
		},
	}

	sgOut, err := c.ec2Svc.CreateSecurityGroup(ctx, sgIn)
	if err != nil {
		return "", errs.FromAwsApi(err, "ec2 create-security-group")
	}

	groupId := aws.ToString(sgOut.GroupId)
	c.logger.DebugContext(ctx, fmt.Sprintf("creaetd %s security group with %s id", in.Name, groupId))

	// add inbound rule to security group

	sgRuleIn := &ec2.AuthorizeSecurityGroupIngressInput{
		CidrIp:     aws.String(inboundCidr),
		GroupId:    aws.String(groupId),
		IpProtocol: aws.String("udp"),
		FromPort:   aws.Int32(int32(inboundPort)),
		ToPort:     aws.Int32(int32(inboundPort)),
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceTypeSecurityGroupRule,
				Tags:         in.toTags(),
			},
		},
	}

	if _, err := c.ec2Svc.AuthorizeSecurityGroupIngress(ctx, sgRuleIn); err != nil {
		return "", errs.FromAwsApi(err, "ec2 authorize-security-group-ingress")
	}
	c.logger.DebugContext(ctx, fmt.Sprintf("added inbound rule to %s security group", groupId))
	return groupId, nil
}

// ListOptedInRegions returns list of opted in regions and default region set in AWS config (or empty string)
func ListOptedInRegions(ctx context.Context, logger *slog.Logger) (Regions, string, error) {
	logger = logger.With("component", "aws.client")
	cfg, err := newAWSConfig("")
	if err != nil {
		return nil, "", err
	}

	out, err := ec2.NewFromConfig(cfg).DescribeRegions(ctx, &ec2.DescribeRegionsInput{})
	if err != nil {
		return nil, "", errs.FromAwsApi(err, "ec2 describe-regions")
	}

	regions := toRegions(out.Regions)
	logger.DebugContext(ctx, fmt.Sprintf("found %d regions, default region set to %q", len(regions), cfg.Region))
	return regions, cfg.Region, nil
}

func newAWSConfig(profile string) (aws.Config, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if profile == "" {
		return config.LoadDefaultConfig(ctx)
	}
	return config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
}
