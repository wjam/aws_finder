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

func vpcCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "vpc [needle]",
		Short: "Find a VPC with the given CIDR range",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return finder.SearchPerRegion(
				cmd.Context(),
				func(ctx context.Context, conf aws.Config) error {
					return findVpc(ctx, args[0], ec2.NewFromConfig(conf))
				})
		},
	}
}

func findVpc(ctx context.Context, needle string, client ec2.DescribeVpcsAPIClient) error {
	pages := ec2.NewDescribeVpcsPaginator(client, nil)

	seq := paginatorToSeq(ctx, pages, vpcsToVpc)
	seq = filter2(func(vpc types.Vpc, err error) bool {
		return err != nil || strings.Contains(*vpc.CidrBlock, needle)
	}, seq)

	for vpc, err := range seq {
		if err != nil {
			return err
		}
		log.Logger(ctx).InfoContext(ctx, aws.ToString(vpc.VpcId))
	}

	return nil
}

func vpcsToVpc(r *ec2.DescribeVpcsOutput) iter.Seq[types.Vpc] {
	return slices.Values(r.Vpcs)
}
