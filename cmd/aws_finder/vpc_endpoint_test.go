package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/stretchr/testify/assert"
)

func TestFindVpcEndpoints(t *testing.T) {
	var tests = []struct {
		endpoints [][]*ec2.VpcEndpoint
		needle    string
		expected  string
	}{
		{
			[][]*ec2.VpcEndpoint{
				{
					{
						VpcEndpointId: aws.String("not used"),
						OwnerId:       aws.String("someone else"),
					},
				},
				{
					{
						VpcEndpointId: aws.String("unused"),
						OwnerId:       aws.String("another owner"),
					},
					{
						VpcEndpointId: aws.String("expected"),
						OwnerId:       aws.String("owner-id"),
					},
				},
			},
			"owner-id",
			"expected",
		},
		{
			[][]*ec2.VpcEndpoint{
				{
					{
						VpcEndpointId: aws.String("not used"),
						ServiceName:   aws.String("someone else"),
					},
				},
				{
					{
						VpcEndpointId: aws.String("unused"),
						ServiceName:   aws.String("another owner"),
					},
					{
						VpcEndpointId: aws.String("expected"),
						ServiceName:   aws.String("service-name"),
					},
				},
			},
			"service-name",
			"expected",
		},
		{
			[][]*ec2.VpcEndpoint{
				{
					{
						VpcEndpointId: aws.String("unused"),
						DnsEntries: []*ec2.DnsEntry{
							{
								DnsName: aws.String("example.org"),
							},
						},
					},
					{
						VpcEndpointId: aws.String("expected"),
						DnsEntries: []*ec2.DnsEntry{
							{
								DnsName: aws.String("example.com"),
							},
							{
								DnsName: aws.String("dns-entry"),
							},
						},
					},
				},
			},
			"dns-entry",
			"expected",
		},
	}

	for _, test := range tests {
		t.Run(test.needle, func(t *testing.T) {
			var buf bytes.Buffer
			findVpcEndpoints(context.TODO(), test.needle, log.New(&buf, "", 0), &vpcEndpointLister{test.endpoints})
			assert.Equal(t, fmt.Sprintf("%s\n", test.expected), buf.String())
		})
	}
}

var _ vpcEndpointPagination = &vpcEndpointLister{}

type vpcEndpointLister struct {
	endpoints [][]*ec2.VpcEndpoint
}

func (v *vpcEndpointLister) DescribeVpcEndpointsPagesWithContext(ctx aws.Context, _ *ec2.DescribeVpcEndpointsInput, fn func(*ec2.DescribeVpcEndpointsOutput, bool) bool, _ ...request.Option) error {
	if ctx == nil {
		return fmt.Errorf("missing context")
	}
	for _, e := range v.endpoints {
		if !fn(&ec2.DescribeVpcEndpointsOutput{
			VpcEndpoints: e,
		}, true) {
			return fmt.Errorf("expected to search all instances")
		}
	}

	return nil
}
