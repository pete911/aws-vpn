package vpc

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"strings"
)

type RouteTables []RouteTable

func toRouteTables(in []types.RouteTable) RouteTables {
	var out RouteTables
	for _, v := range in {
		out = append(out, toRouteTable(v))
	}
	return out
}

func (r RouteTables) GetBySubnet(vpcId, subnetId string) (RouteTable, bool) {
	for _, rtb := range r.getByVpcId(vpcId) {
		for _, id := range rtb.SubnetIds {
			if id == subnetId {
				return rtb, true
			}
		}
	}
	// no association found, return main route table for VPC
	return r.getMain(vpcId)
}

func (r RouteTables) getByVpcId(id string) RouteTables {
	var out RouteTables
	for _, rtb := range r {
		if rtb.VpcId == id {
			out = append(out, rtb)
		}
	}
	return out
}

func (r RouteTables) getMain(vpcId string) (RouteTable, bool) {
	for _, rt := range r {
		if rt.Main && rt.VpcId == vpcId {
			return rt, true
		}
	}
	return RouteTable{}, false
}

type RouteTable struct {
	Id        string
	VpcId     string
	Main      bool
	SubnetIds []string
	OwnerId   string
	Name      string
	Routes    Routes
}

func toRouteTable(in types.RouteTable) RouteTable {
	var routes Routes
	for _, v := range in.Routes {
		routes = append(routes, toRoute(v))
	}

	var main bool
	var subnetIds []string
	for _, association := range in.Associations {
		if ok := aws.ToBool(association.Main); ok {
			main = true
		}
		if subnetId := aws.ToString(association.SubnetId); subnetId != "" {
			subnetIds = append(subnetIds, subnetId)
		}
	}

	return RouteTable{
		Id:        aws.ToString(in.RouteTableId),
		VpcId:     aws.ToString(in.VpcId),
		Main:      main,
		SubnetIds: subnetIds,
		OwnerId:   aws.ToString(in.OwnerId),
		Name:      fromTags(in.Tags)["Name"],
		Routes:    routes,
	}
}

// HasPublicRoute returns true if there is 0.0.0.0/0 IPv4 destination mapped to IGW
func (r RouteTable) HasPublicRoute() bool {
	return r.Routes.hasPublicRoute()
}

type Routes []Route

func (r Routes) hasPublicRoute() bool {
	for _, route := range r {
		if route.DestinationType == "ipv4" && route.DestinationCidr == "0.0.0.0/0" {
			return route.TargetType == "internet-gateway"
		}
	}
	return false
}

type Route struct {
	DestinationType string
	DestinationCidr string
	TargetId        string
	TargetType      string
	State           string
}

func toRoute(in types.Route) Route {
	var route Route
	route.State = string(in.State)

	// destination
	if id := aws.ToString(in.DestinationPrefixListId); id != "" {
		route.DestinationType = "prefix-list"
	}
	if cidr := aws.ToString(in.DestinationCidrBlock); cidr != "" {
		route.DestinationType = "ipv4"
		route.DestinationCidr = cidr
	}
	if cidr := aws.ToString(in.DestinationIpv6CidrBlock); cidr != "" {
		route.DestinationType = "ipv6"
		route.DestinationCidr = cidr
	}

	// target
	if id := aws.ToString(in.CarrierGatewayId); id != "" {
		route.TargetId = id
		route.TargetType = "carrier-gateway"
		return route
	}
	if id := aws.ToString(in.CoreNetworkArn); id != "" {
		route.TargetId = id
		route.TargetType = "core-network"
		return route
	}
	if id := aws.ToString(in.EgressOnlyInternetGatewayId); id != "" {
		route.TargetId = id
		route.TargetType = "egress-only-internet-gateway"
		return route
	}

	if id := aws.ToString(in.GatewayId); id != "" {
		route.TargetId = id
		if id == "local" {
			route.TargetType = id
			return route
		}
		if strings.HasPrefix(id, "igw-") {
			route.TargetType = "internet-gateway"
			return route
		}
		if strings.HasPrefix(id, "vpce-") {
			route.TargetType = "vpc-endpoint"
			return route
		}
	}

	if id := aws.ToString(in.InstanceId); id != "" {
		route.TargetId = id
		route.TargetType = "nat-instance"
		return route
	}
	if id := aws.ToString(in.LocalGatewayId); id != "" {
		route.TargetId = id
		route.TargetType = "local-gateway"
		return route
	}
	if id := aws.ToString(in.NatGatewayId); id != "" {
		route.TargetId = id
		route.TargetType = "nat-gateway"
		return route
	}
	if id := aws.ToString(in.NetworkInterfaceId); id != "" {
		route.TargetId = id
		route.TargetType = "network-interface"
		return route
	}
	if id := aws.ToString(in.TransitGatewayId); id != "" {
		route.TargetId = id
		route.TargetType = "transit-gateway"
		return route
	}
	if id := aws.ToString(in.VpcPeeringConnectionId); id != "" {
		route.TargetId = id
		route.TargetType = "vpc-peering-connection"
		return route
	}
	return route
}
