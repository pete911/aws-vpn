package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/pete911/aws-vpn/internal/aws/iam"
	"strings"
	"time"
)

type MetadataInput struct {
	Name string
	Tags map[string]string
}

func (m MetadataInput) toTagFilter() []types.Filter {
	var out []types.Filter
	for k, v := range m.Tags {
		out = append(out, types.Filter{Name: aws.String(fmt.Sprintf("tag:%s", k)), Values: []string{v}})
	}
	return out
}

func (m MetadataInput) toTags() []types.Tag {
	var out []types.Tag
	for k, v := range m.Tags {
		out = append(out, types.Tag{Key: aws.String(k), Value: aws.String(v)})
	}
	return out
}

type RunInstancesInput struct {
	Metadata        MetadataInput
	SubnetId        string
	InboundCidr     string
	InboundPort     int
	UserData        string
	SecretsPath     string
	InstanceProfile iam.InstanceProfileInput
}

type InstanceStatus struct {
	InstanceId     string
	InstanceState  string
	InstanceStatus string
	SystemStatus   string
	EbsStatus      string
}

func ToInstanceStatus(in types.InstanceStatus) InstanceStatus {
	var state, status, systemStatus, ebsStatus string
	if in.InstanceState != nil {
		state = string(in.InstanceState.Name)
	}
	if in.InstanceStatus != nil {
		status = string(in.InstanceStatus.Status)
	}
	if in.SystemStatus != nil {
		systemStatus = string(in.SystemStatus.Status)
	}
	if in.AttachedEbsStatus != nil {
		ebsStatus = string(in.AttachedEbsStatus.Status)
	}

	return InstanceStatus{
		InstanceId:     aws.ToString(in.InstanceId),
		InstanceState:  state,
		InstanceStatus: status,
		SystemStatus:   systemStatus,
		EbsStatus:      ebsStatus,
	}
}

func (i InstanceStatus) IsReady() bool {
	if i.InstanceState != "running" {
		return false
	}
	if i.InstanceStatus != "ok" {
		return false
	}
	if i.SystemStatus != "ok" {
		return false
	}
	if i.EbsStatus != "" && i.EbsStatus != "ok" {
		return false
	}
	return true
}

func (i InstanceStatus) String() string {
	return fmt.Sprintf("state: %s, status: %s, system-status: %s, ebs-status: %s",
		i.InstanceState, i.InstanceStatus, i.SystemStatus, i.EbsStatus)
}

type Instances []Instance

func (i Instances) Names() []string {
	var out []string
	for _, instance := range i {
		out = append(out, instance.Name)
	}
	return out
}

type Instance struct {
	Id                 string
	Name               string
	InstanceProfile    string
	SecurityGroupNames []string
	PublicDnsName      string
	PublicIp           string
	PrivateDnsName     string
	PrivateIp          string
	ImageId            string
	InstanceType       string
	State              string
	StateReason        string
	LaunchTime         time.Time
	Tags               map[string]string
}

func ToInstances(in []types.Instance) []Instance {
	var out []Instance
	for _, v := range in {
		out = append(out, ToInstance(v))
	}
	return out
}

func ToInstance(in types.Instance) Instance {
	var instanceProfile string
	if in.IamInstanceProfile != nil {
		if arnParts := strings.Split(aws.ToString(in.IamInstanceProfile.Arn), "/"); len(arnParts) == 2 {
			instanceProfile = arnParts[1]
		}
	}
	var securityGroupNames []string
	for _, v := range in.SecurityGroups {
		securityGroupNames = append(securityGroupNames, aws.ToString(v.GroupName))
	}
	var state, stateReason string
	if in.State != nil {
		state = string(in.State.Name)
	}
	if in.StateReason != nil {
		stateReason = aws.ToString(in.StateReason.Message)
	}
	tags := fromTags(in.Tags)

	return Instance{
		Id:                 aws.ToString(in.InstanceId),
		Name:               tags["Name"],
		InstanceProfile:    instanceProfile,
		SecurityGroupNames: securityGroupNames,
		PublicDnsName:      aws.ToString(in.PublicDnsName),
		PublicIp:           aws.ToString(in.PublicIpAddress),
		PrivateDnsName:     aws.ToString(in.PrivateDnsName),
		PrivateIp:          aws.ToString(in.PrivateIpAddress),
		ImageId:            aws.ToString(in.ImageId),
		InstanceType:       string(in.InstanceType),
		State:              state,
		StateReason:        stateReason,
		LaunchTime:         aws.ToTime(in.LaunchTime),
		Tags:               tags,
	}
}

func fromTags(in []types.Tag) map[string]string {
	var out = make(map[string]string)
	for _, v := range in {
		out[aws.ToString(v.Key)] = aws.ToString(v.Value)
	}
	return out
}
