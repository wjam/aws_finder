package main

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/cobra"
	"github.com/wjam/aws_finder/internal/finder"
)

func init() {
	commands = append(commands, &cobra.Command{
		Use:   "vpc_endpoint [needle]",
		Short: "Find a VPC endpoint by the given service name",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			finder.SearchPerRegion(cmd.Context(), func(ctx context.Context, l *log.Logger, sess *session.Session) {
				findVpcEndpoints(ctx, args[0], l, ec2.New(sess))
			})
		},
	})
}

func findVpcEndpoints(ctx context.Context, needle string, l *log.Logger, client vpcEndpointPagination) {
	err := client.DescribeVpcEndpointsPagesWithContext(ctx, &ec2.DescribeVpcEndpointsInput{}, func(output *ec2.DescribeVpcEndpointsOutput, more bool) bool {
		for _, endpoint := range output.VpcEndpoints {
			if findVpcEndpoint(needle, endpoint) {
				l.Println(aws.StringValue(endpoint.VpcEndpointId))
			}
		}
		return true
	})
	if err != nil {
		l.Printf("Failed to query endpoints: %s", err)
		return
	}
}

func findVpcEndpoint(needle string, endpoint *ec2.VpcEndpoint) bool {
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

type vpcEndpointPagination interface {
	DescribeVpcEndpointsPagesWithContext(ctx aws.Context, input *ec2.DescribeVpcEndpointsInput, fn func(*ec2.DescribeVpcEndpointsOutput, bool) bool, opts ...request.Option) error
}
