package vpc

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/pete911/aws-vpn/internal/errs"
	"log/slog"
)

type Service struct {
	logger *slog.Logger
	svc    *ec2.Client
}

func NewService(logger *slog.Logger, cfg aws.Config) Service {
	return Service{
		logger: logger.With("component", "aws.vpc.service"),
		svc:    ec2.NewFromConfig(cfg),
	}
}

// GetDefaultPublicSubnets returns public subnets from default VPC
func (s Service) GetDefaultPublicSubnets(ctx context.Context) (Subnets, error) {
	vpcId, err := s.getDefaultVpcId(ctx)
	if err != nil {
		return nil, err
	}
	routeTables, err := s.describeRouteTables(ctx, vpcId)
	if err != nil {
		return nil, err
	}
	subnets, err := s.describeSubnets(ctx, vpcId)
	if err != nil {
		return nil, err
	}

	// find public subnets
	var publicSubnets Subnets
	for _, subnet := range subnets {
		// find route table for this subnet, if route table is found and has public route (IGW for 0.0.0.0/0) add it
		if rtb, ok := routeTables.GetBySubnet(vpcId, subnet.Id); ok && rtb.HasPublicRoute() {
			publicSubnets = append(publicSubnets, subnet)
		}
	}
	s.logger.DebugContext(ctx, fmt.Sprintf("found %d public subnets in vpc %s", len(publicSubnets), vpcId))
	return publicSubnets, nil
}

func (s Service) describeSubnets(ctx context.Context, vpcId string) (Subnets, error) {
	in := &ec2.DescribeSubnetsInput{
		Filters: []ec2types.Filter{
			{Name: aws.String("vpc-id"), Values: []string{vpcId}},
			{Name: aws.String("state"), Values: []string{"available"}},
		},
	}

	var subnets Subnets
	for {
		out, err := s.svc.DescribeSubnets(ctx, in)
		if err != nil {
			return nil, errs.FromAwsApi(err, "ec2 describe-subnets")
		}
		subnets = append(subnets, toSubnets(out.Subnets)...)

		if aws.ToString(out.NextToken) == "" {
			break
		}
		in.NextToken = out.NextToken
	}
	s.logger.DebugContext(ctx, fmt.Sprintf("found %d subnets in vpc %s", len(subnets), vpcId))
	return subnets, nil
}

func (s Service) describeRouteTables(ctx context.Context, vpcId string) (RouteTables, error) {
	in := &ec2.DescribeRouteTablesInput{
		DryRun: nil,
		Filters: []ec2types.Filter{
			{Name: aws.String("vpc-id"), Values: []string{vpcId}},
		},
	}

	var routeTables RouteTables
	for {
		out, err := s.svc.DescribeRouteTables(ctx, in)
		if err != nil {
			return nil, errs.FromAwsApi(err, "ec2 describe-route-tables")
		}
		routeTables = append(routeTables, toRouteTables(out.RouteTables)...)

		if aws.ToString(out.NextToken) == "" {
			break
		}
		in.NextToken = out.NextToken
	}
	s.logger.DebugContext(ctx, fmt.Sprintf("found %d route tables in vpc %s", len(routeTables), vpcId))
	return routeTables, nil
}

func (s Service) getDefaultVpcId(ctx context.Context) (string, error) {
	in := &ec2.DescribeVpcsInput{
		Filters: []ec2types.Filter{
			{Name: aws.String("is-default"), Values: []string{"true"}},
			{Name: aws.String("state"), Values: []string{"available"}},
		},
	}

	// no need to paginate, we just want default vpc
	out, err := s.svc.DescribeVpcs(ctx, in)
	if err != nil {
		return "", errs.FromAwsApi(err, "ec2 describe-vpcs")
	}
	if out == nil || len(out.Vpcs) == 0 {
		return "", errors.New("no default vpc found")
	}

	vpcId := aws.ToString(out.Vpcs[0].VpcId)
	vpcName := fromTags(out.Vpcs[0].Tags)["Name"]
	s.logger.DebugContext(ctx, fmt.Sprintf("found default vpc %s %s", vpcId, vpcName))
	return vpcId, nil
}
