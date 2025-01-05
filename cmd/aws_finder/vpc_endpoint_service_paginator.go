package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

type describeVpcEndpointServicesClient interface {
	DescribeVpcEndpointServices(ctx context.Context, params *ec2.DescribeVpcEndpointServicesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVpcEndpointServicesOutput, error)
}

type describeVpcEndpointServicesPaginator struct {
	client    describeVpcEndpointServicesClient
	params    *ec2.DescribeVpcEndpointServicesInput
	nextToken *string
	firstPage bool
}

func newDescribeVpcEndpointServicesPaginator(client describeVpcEndpointServicesClient, params *ec2.DescribeVpcEndpointServicesInput) *describeVpcEndpointServicesPaginator {
	if params == nil {
		params = &ec2.DescribeVpcEndpointServicesInput{}
	}
	return &describeVpcEndpointServicesPaginator{
		client:    client,
		params:    params,
		firstPage: true,
	}
}

func (p *describeVpcEndpointServicesPaginator) HasMorePages() bool {
	return p.firstPage || p.nextToken != nil
}

func (p *describeVpcEndpointServicesPaginator) NextPage(ctx context.Context, _ ...func(string)) (*ec2.DescribeVpcEndpointServicesOutput, error) {
	if !p.HasMorePages() {
		return nil, fmt.Errorf("no more pages available")
	}

	params := *p.params
	params.NextToken = p.nextToken

	result, err := p.client.DescribeVpcEndpointServices(ctx, &params)
	if err != nil {
		return nil, err
	}
	p.firstPage = false

	p.nextToken = result.NextToken

	return result, nil
}
