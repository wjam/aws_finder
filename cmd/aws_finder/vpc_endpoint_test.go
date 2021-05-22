package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/stretchr/testify/assert"
)

func TestFindVpcEndpoints(t *testing.T) {
	var tests = []struct {
		endpoints [][]types.VpcEndpoint
		needle    string
		expected  string
	}{
		{
			[][]types.VpcEndpoint{
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
			[][]types.VpcEndpoint{
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
			[][]types.VpcEndpoint{
				{
					{
						VpcEndpointId: aws.String("unused"),
						DnsEntries: []types.DnsEntry{
							{
								DnsName: aws.String("example.org"),
							},
						},
					},
					{
						VpcEndpointId: aws.String("expected"),
						DnsEntries: []types.DnsEntry{
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

var _ ec2.DescribeVpcEndpointsAPIClient = &vpcEndpointLister{}

type vpcEndpointLister struct {
	endpoints [][]types.VpcEndpoint
}

func (v *vpcEndpointLister) DescribeVpcEndpoints(ctx context.Context, _ *ec2.DescribeVpcEndpointsInput, _ ...func(*ec2.Options)) (*ec2.DescribeVpcEndpointsOutput, error) {
	if ctx == nil {
		return nil, fmt.Errorf("missing context")
	}
	if len(v.endpoints) == 0 {
		return nil, fmt.Errorf("no more values")
	}

	var value []types.VpcEndpoint
	value, v.endpoints = v.endpoints[0], v.endpoints[1:]

	var token *string
	if len(v.endpoints) != 0 {
		token = aws.String(strconv.Itoa(len(v.endpoints)))
	}

	return &ec2.DescribeVpcEndpointsOutput{
		NextToken:    token,
		VpcEndpoints: value,
	}, nil
}
