package main

import (
	"context"
	"iter"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/spf13/cobra"
	"github.com/wjam/aws_finder/internal/finder"
	"github.com/wjam/aws_finder/internal/log"
)

func vpcEndpointCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "vpc_endpoint [needle]",
		Short: "Find a VPC endpoint by the given service name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return finder.SearchPerRegion(
				cmd.Context(),
				func(ctx context.Context, conf aws.Config) error {
					return findVpcEndpoints(ctx, args[0], ec2.NewFromConfig(conf))
				})
		},
	}
}

func findVpcEndpoints(
	ctx context.Context, needle string, client ec2.DescribeVpcEndpointsAPIClient,
) error {
	pages := ec2.NewDescribeVpcEndpointsPaginator(client, nil)

	seq := paginatorToSeq(ctx, pages, vpcEndpointsToVpcEndpoint)
	seq = filter2(func(endpoint types.VpcEndpoint, err error) bool {
		return err != nil || findVpcEndpoint(needle, endpoint)
	}, seq)

	for endpoint, err := range seq {
		if err != nil {
			return err
		}
		log.Logger(ctx).InfoContext(ctx, aws.ToString(endpoint.VpcEndpointId))
	}

	return nil
}

func vpcEndpointsToVpcEndpoint(r *ec2.DescribeVpcEndpointsOutput) iter.Seq[types.VpcEndpoint] {
	return slices.Values(r.VpcEndpoints)
}

func findVpcEndpoint(needle string, endpoint types.VpcEndpoint) bool {
	if check(needle, endpoint.OwnerId, endpoint.ServiceName) {
		return true
	}
	for _, entry := range endpoint.DnsEntries {
		if check(needle, entry.DnsName) {
			return true
		}
	}
	return false
}
