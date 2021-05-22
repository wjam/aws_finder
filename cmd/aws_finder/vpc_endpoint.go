package main

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/spf13/cobra"
	"github.com/wjam/aws_finder/internal/finder"
)

func init() {
	commands = append(commands, &cobra.Command{
		Use:   "vpc_endpoint [needle]",
		Short: "Find a VPC endpoint by the given service name",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			finder.SearchPerRegion(cmd.Context(), func(ctx context.Context, l *log.Logger, conf aws.Config) {
				findVpcEndpoints(ctx, args[0], l, ec2.NewFromConfig(conf))
			})
		},
	})
}

func findVpcEndpoints(ctx context.Context, needle string, l *log.Logger, client ec2.DescribeVpcEndpointsAPIClient) {
	pages := ec2.NewDescribeVpcEndpointsPaginator(client, &ec2.DescribeVpcEndpointsInput{})

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			l.Printf("Failed to query endpoints: %s", err)
			return
		}

		for _, endpoint := range page.VpcEndpoints {
			if findVpcEndpoint(needle, endpoint) {
				l.Println(aws.ToString(endpoint.VpcEndpointId))
			}
		}
	}
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
