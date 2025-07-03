package iam

import (
	"fmt"
	"strings"
)

var (
	ec2AssumeRolePolicyDocument = `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "Service": "ec2.amazonaws.com"
            },
            "Action": "sts:AssumeRole"
        }
    ]
}`
)

type InlinePolicyInput struct {
	Name     string
	Document string
}

func NewInlinePolicyInput(name, resource string, actions []string) InlinePolicyInput {
	return InlinePolicyInput{
		Name:     name,
		Document: newDocument(resource, actions),
	}
}

func newDocument(resource string, actions []string) string {
	if resource != "" {
		resource = "*"
	}
	actionsList := "*"
	if len(actions) > 0 {
		actionsList = strings.Join(actions, `", "`)
	}

	return fmt.Sprintf(`{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": ["%s"],
            "Effect": "Allow",
            "Resource": "%s",
            "Sid": "VPNSecrets"
        }
    ]
}`, actionsList, resource)
}
