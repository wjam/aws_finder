package main

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"

	"github.com/spf13/cobra"
	"github.com/wjam/aws_finder/internal/finder"
)

func init() {
	commands = append(commands, &cobra.Command{
		Use:   "vpc [needle]",
		Short: "Find a VPC with the given CIDR range",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			finder.SearchPerRegion(cmd.Context(), func(ctx context.Context, l *log.Logger, conf aws.Config) {
				findVpc(ctx, args[0], l, ec2.NewFromConfig(conf))
			})
		},
	})
}

func findVpc(ctx context.Context, needle string, l *log.Logger, client vpcLister) {
	pages := ec2.NewDescribeVpcsPaginator(client, &ec2.DescribeVpcsInput{})

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			l.Printf("Failed to query VPCs: %s", err)
			return
		}

		for _, vpc := range page.Vpcs {
			if strings.Contains(*vpc.CidrBlock, needle) {
				l.Println(*vpc.VpcId)
			}
		}
	}
}

type vpcLister interface {
	DescribeVpcs(context.Context, *ec2.DescribeVpcsInput, ...func(*ec2.Options)) (*ec2.DescribeVpcsOutput, error)
}
