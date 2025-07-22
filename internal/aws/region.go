package aws

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"sort"
)

var regionsMap = map[string]Region{
	"us-east-1":      {Code: "us-east-1", Name: "US East (N. Virginia)", Geography: "United States of America"},
	"us-east-2":      {Code: "us-east-2", Name: "US East (Ohio)", Geography: "United States of America"},
	"us-west-1":      {Code: "us-west-1", Name: "US West (N. California)", Geography: "United States of America"},
	"us-west-2":      {Code: "us-west-2", Name: "US West (Oregon)", Geography: "United States of America"},
	"af-south-1":     {Code: "af-south-1", Name: "Africa (Cape Town)", Geography: "South Africa"},
	"ap-east-1":      {Code: "ap-east-1", Name: "Asia Pacific (Hong Kong)", Geography: "Hong Kong"},
	"ap-south-2":     {Code: "ap-south-2", Name: "Asia Pacific (Hyderabad)", Geography: "India"},
	"ap-southeast-3": {Code: "ap-southeast-3", Name: "Asia Pacific (Jakarta)", Geography: "Indonesia"},
	"ap-southeast-5": {Code: "ap-southeast-5", Name: "Asia Pacific (Malaysia)", Geography: "Malaysia"},
	"ap-southeast-4": {Code: "ap-southeast-4", Name: "Asia Pacific (Melbourne)", Geography: "Australia"},
	"ap-south-1":     {Code: "ap-south-1", Name: "Asia Pacific (Mumbai)", Geography: "India"},
	"ap-northeast-3": {Code: "ap-northeast-3", Name: "Asia Pacific (Osaka)", Geography: "Japan"},
	"ap-northeast-2": {Code: "ap-northeast-2", Name: "Asia Pacific (Seoul)", Geography: "South Korea"},
	"ap-southeast-1": {Code: "ap-southeast-1", Name: "Asia Pacific (Singapore)", Geography: "Singapore"},
	"ap-southeast-2": {Code: "ap-southeast-2", Name: "Asia Pacific (Sydney)", Geography: "Australia"},
	"ap-southeast-7": {Code: "ap-southeast-7", Name: "Asia Pacific (Thailand)", Geography: "Thailand"},
	"ap-northeast-1": {Code: "ap-northeast-1", Name: "Asia Pacific (Tokyo)", Geography: "Japan"},
	"ca-central-1":   {Code: "ca-central-1", Name: "Canada (Central)", Geography: "Canada"},
	"ca-west-1":      {Code: "ca-west-1", Name: "Canada West (Calgary)", Geography: "Canada"},
	"eu-central-1":   {Code: "eu-central-1", Name: "Europe (Frankfurt)", Geography: "Germany"},
	"eu-west-1":      {Code: "eu-west-1", Name: "Europe (Ireland)", Geography: "Ireland"},
	"eu-west-2":      {Code: "eu-west-2", Name: "Europe (London)", Geography: "United Kingdom"},
	"eu-south-1":     {Code: "eu-south-1", Name: "Europe (Milan)", Geography: "Italy"},
	"eu-west-3":      {Code: "eu-west-3", Name: "Europe (Paris)", Geography: "France"},
	"eu-south-2":     {Code: "eu-south-2", Name: "Europe (Spain)", Geography: "Spain"},
	"eu-north-1":     {Code: "eu-north-1", Name: "Europe (Stockholm)", Geography: "Sweden"},
	"eu-central-2":   {Code: "eu-central-2", Name: "Europe (Zurich)", Geography: "Switzerland"},
	"il-central-1":   {Code: "il-central-1", Name: "Israel (Tel Aviv)", Geography: "Israel"},
	"mx-central-1":   {Code: "mx-central-1", Name: "Mexico (Central)", Geography: "Mexico"},
	"me-south-1":     {Code: "me-south-1", Name: "Middle East (Bahrain)", Geography: "Bahrain"},
	"me-central-1":   {Code: "me-central-1", Name: "Middle East (UAE)", Geography: "United Arab Emirates"},
	"sa-east-1":      {Code: "sa-east-1", Name: "South America (São Paulo)", Geography: "Brazil"},
}

type Regions []Region

func toRegions(in []types.Region) Regions {
	var out Regions
	for _, region := range in {
		out = append(out, regionsMap[aws.ToString(region.RegionName)])
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})
	return out
}

func (r Regions) Names() []string {
	var out []string
	for _, region := range r {
		out = append(out, region.Name)
	}
	return out
}

func RegionByCode(code string) Region {
	return regionsMap[code]
}

type Region struct {
	Code      string
	Name      string
	Geography string
}
