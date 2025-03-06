package main

import (
	"context"
	"iter"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/spf13/cobra"
	"github.com/wjam/aws_finder/internal/finder"
	"github.com/wjam/aws_finder/internal/log"
)

func vpcEndpointServiceCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "vpc_endpoint_service [needle]",
		Short: "Find a VPC endpoint service by the given service name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return finder.SearchPerRegion(
				cmd.Context(),
				func(ctx context.Context, conf aws.Config) error {
					return findVpcEndpointService(ctx, args[0], ec2.NewFromConfig(conf))
				})
		},
	}
}

func findVpcEndpointService(
	ctx context.Context, needle string, client describeVpcEndpointServicesClient,
) error {
	pages := newDescribeVpcEndpointServicesPaginator(client, nil)

	seq := paginatorToSeq(ctx, pages, vpcEndpointServicesToServiceDetail)
	seq = filter2(func(svc types.ServiceDetail, err error) bool {
		return err != nil || strings.Contains(*svc.ServiceName, needle)
	}, seq)

	for svc, err := range seq {
		if err != nil {
			return err
		}

		log.Logger(ctx).InfoContext(ctx, aws.ToString(svc.ServiceName))
	}

	return nil
}

func vpcEndpointServicesToServiceDetail(r *ec2.DescribeVpcEndpointServicesOutput) iter.Seq[types.ServiceDetail] {
	return slices.Values(r.ServiceDetails)
}
