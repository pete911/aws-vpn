package vpc

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func fromTags(in []types.Tag) map[string]string {
	var out = make(map[string]string)
	for _, v := range in {
		out[aws.ToString(v.Key)] = aws.ToString(v.Value)
	}
	return out
}
