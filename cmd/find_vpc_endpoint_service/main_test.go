package main

import (
	"bytes"
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/stretchr/testify/assert"
)

func TestSearch(t *testing.T) {
	var buf bytes.Buffer
	search("find", log.New(&buf, "", 0), &vpcEndpoints{
		data: map[string]ec2.DescribeVpcEndpointServicesOutput{
			"": {
				NextToken: aws.String("next-one"),
				ServiceDetails: []*ec2.ServiceDetail{
					{
						ServiceName: aws.String("first one"),
					},
					{
						ServiceName: aws.String("second"),
					},
					{
						ServiceName: aws.String("another"),
					},
				},
			},
			"next-one": {
				NextToken: nil,
				ServiceDetails: []*ec2.ServiceDetail{
					{
						ServiceName: aws.String("ignored"),
					},
					{
						ServiceName: aws.String("one to find"),
					},
					{
						ServiceName: aws.String("glossed over"),
					},
				},
			},
		},
	})

	assert.Equal(t, "one to find\n", buf.String())
}

var _ vpcEndpointLister = &vpcEndpoints{}

type vpcEndpoints struct {
	data map[string]ec2.DescribeVpcEndpointServicesOutput
}

func (v *vpcEndpoints) DescribeVpcEndpointServices(input *ec2.DescribeVpcEndpointServicesInput) (*ec2.DescribeVpcEndpointServicesOutput, error) {
	if data, ok := v.data[aws.StringValue(input.NextToken)]; ok {
		return &data, nil
	}
	return nil, fmt.Errorf("unknown key %s", *input.NextToken)
}
