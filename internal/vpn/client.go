package vpn

import (
	"context"
	"errors"
	"fmt"
	"github.com/pete911/aws-vpn/internal/aws"
	"github.com/pete911/aws-vpn/internal/aws/iam"
	"github.com/pete911/aws-vpn/internal/aws/vpc"
	"github.com/pete911/aws-vpn/internal/vpn/ovpn"
	"github.com/pete911/aws-vpn/internal/vpn/wg"
	"log/slog"
	"time"
)

const NamePrefix = "aws-vpn-"

type ProductType int

const (
	OpenVpn ProductType = iota
	WireGuard
)

func getMetadataInput(name string) aws.MetadataInput {
	name = NamePrefix + name
	return aws.MetadataInput{
		Name: name,
		Tags: map[string]string{
			"Name":       name,
			"Project":    "aws-vpn",
			"Repository": "https://github.com/pete911/aws-vpn",
		},
	}
}

type Client struct {
	Region      string
	logger      *slog.Logger
	awsClient   aws.Client
	productType ProductType
}

func NewClient(logger *slog.Logger, awsClient aws.Client, product ProductType) Client {
	return Client{
		Region:      awsClient.Region,
		logger:      logger.With("component", "vpn.client"),
		awsClient:   awsClient,
		productType: product,
	}
}

func (c Client) Delete(instance aws.Instance) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*300)
	defer cancel()

	return c.awsClient.TerminateInstance(ctx, instance, GetSecretsPath(instance.Name))
}

func (c Client) List() (aws.Instances, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// we don't care about name in the tags (it will be stripped anyway), so providing just empty string to get tags
	return c.awsClient.DescribeInstancesByNamePrefix(ctx, NamePrefix, getMetadataInput("").Tags)
}

func (c Client) GetClientConfig(instance aws.Instance) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	prefix := GetSecretsPath(instance.Name)
	secrets, err := c.awsClient.GetSecrets(ctx, prefix)
	if err != nil {
		return nil, err
	}

	clientConfig := ovpn.NewClientConfig(instance.PublicDnsName, secrets)
	return clientConfig.Parse()
}

func (c Client) Create(name, inboundCidr string) (aws.Instance, error) {
	subnet, err := c.getSubnet()
	if err != nil {
		return aws.Instance{}, err
	}
	c.logger.Debug(fmt.Sprintf("selected %s %s subnet in %s AZ", subnet.Id, subnet.Name, subnet.AvailabilityZone))

	instance, err := c.runInstance(name, subnet.Id, inboundCidr)
	if err != nil {
		return aws.Instance{}, err
	}
	c.logger.Info(fmt.Sprintf("starting instance %s in subnet %s AZ %s", instance.Id, subnet.Id, subnet.AvailabilityZone))
	c.logger.Info("waiting 60 seconds for instance to initialize")
	time.Sleep(60 * time.Second)

	// wait for instance to start
	for x := 0; x < 30; x++ {
		time.Sleep(15 * time.Second)
		status, err := c.describeInstanceStatus(instance.Id)
		if err != nil {
			return aws.Instance{}, err
		}

		c.logger.Info(fmt.Sprintf("instance %s - %s", instance.Id, status))
		if status.IsReady() {
			// get fresh initialized instance with public IP and dns set
			return c.describeInstanceById(instance.Id)
		}
		c.logger.Info("retry in 15 seconds")
	}
	return aws.Instance{}, fmt.Errorf("instance %s not ready", instance.Id)
}

func (c Client) describeInstanceStatus(id string) (aws.InstanceStatus, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	return c.awsClient.DescribeInstanceStatus(ctx, id)
}

func (c Client) describeInstanceById(id string) (aws.Instance, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	return c.awsClient.DescribeInstanceById(ctx, id)
}

func (c Client) runInstance(name, subnetId, inboundCidr string) (aws.Instance, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	meta := getMetadataInput(name)
	config := NewConfig(meta.Name, c.awsClient.AccountId, c.awsClient.Region)
	userData, err := c.userData(config)
	if err != nil {
		return aws.Instance{}, err
	}
	inboundPort, err := c.inboundPort()
	if err != nil {
		return aws.Instance{}, err
	}

	input := aws.RunInstancesInput{
		Metadata:    meta,
		SubnetId:    subnetId,
		InboundCidr: inboundCidr,
		InboundPort: inboundPort,
		UserData:    userData,
		SecretsPath: config.SecretsPath,
		InstanceProfile: iam.InstanceProfileInput{
			Name: meta.Name,
			Tags: meta.Tags,
			Role: iam.RoleInput{
				RoleName:           fmt.Sprintf("%s-%s", meta.Name, config.Region),
				ManagedPolicyNames: GetSSMManagedPolicies(),
				InlinePolicies:     []iam.InlinePolicyInput{config.GetSecretsInlinePolicy()},
				Tags:               meta.Tags,
			},
		},
	}
	return c.awsClient.RunInstance(ctx, input)
}

func (c Client) userData(config any) (string, error) {
	if c.productType == OpenVpn {
		return ovpn.UserData(config)
	}
	if c.productType == WireGuard {
		return wg.UserData(config)
	}
	return "", fmt.Errorf("unknown vpn product type: %d", c.productType)
}

func (c Client) inboundPort() (int, error) {
	if c.productType == OpenVpn {
		return ovpn.InboundPort, nil
	}
	if c.productType == WireGuard {
		return wg.InboundPort, nil
	}
	return -1, fmt.Errorf("unknown vpn product type: %d", c.productType)
}

func (c Client) getSubnet() (vpc.Subnet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	subnets, err := c.awsClient.GetDefaultPublicSubnets(ctx)
	if err != nil {
		return vpc.Subnet{}, err
	}

	if len(subnets) == 0 {
		return vpc.Subnet{}, errors.New("no default public subnets found")
	}
	return subnets[0], nil
}
