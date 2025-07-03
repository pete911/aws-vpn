package vpc

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type Subnets []Subnet

func toSubnets(in []types.Subnet) Subnets {
	var out Subnets
	for _, v := range in {
		out = append(out, toSubnet(v))
	}
	return out
}

type Subnet struct {
	VpcId                   string
	Id                      string
	SubnetArn               string
	Name                    string
	AvailabilityZone        string
	AvailabilityZoneId      string
	AvailableIpAddressCount int
	CidrBlock               string
	DefaultForAz            bool
	State                   string
}

func toSubnet(in types.Subnet) Subnet {
	return Subnet{
		VpcId:                   aws.ToString(in.VpcId),
		Id:                      aws.ToString(in.SubnetId),
		SubnetArn:               aws.ToString(in.SubnetArn),
		Name:                    fromTags(in.Tags)["Name"],
		AvailabilityZone:        aws.ToString(in.AvailabilityZone),
		AvailabilityZoneId:      aws.ToString(in.AvailabilityZoneId),
		AvailableIpAddressCount: int(aws.ToInt32(in.AvailableIpAddressCount)),
		CidrBlock:               aws.ToString(in.CidrBlock),
		DefaultForAz:            aws.ToBool(in.DefaultForAz),
		State:                   string(in.State),
	}
}
