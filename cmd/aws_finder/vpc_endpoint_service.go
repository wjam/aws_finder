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
		Use:   "vpc_endpoint_service [needle]",
		Short: "Find a VPC endpoint service by the given service name",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			finder.SearchPerRegion(cmd.Context(), func(ctx context.Context, l *log.Logger, conf aws.Config) {
				findVpcEndpointService(ctx, args[0], l, ec2.NewFromConfig(conf))
			})
		},
	})
}

func findVpcEndpointService(ctx context.Context, needle string, l *log.Logger, client vpcEndpointServiceLister) {
	var next *string
	for {
		output, err := client.DescribeVpcEndpointServices(ctx, &ec2.DescribeVpcEndpointServicesInput{
			NextToken: next,
		})
		if err != nil {
			l.Printf("Failed to query vpc endpoint services: %s", err)
			return
		}

		for _, svc := range output.ServiceDetails {
			if strings.Contains(*svc.ServiceName, needle) {
				l.Printf(*svc.ServiceName)
			}
		}

		next = output.NextToken

		if next == nil {
			break
		}
	}
}

type vpcEndpointServiceLister interface {
	DescribeVpcEndpointServices(ctx context.Context, params *ec2.DescribeVpcEndpointServicesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVpcEndpointServicesOutput, error)
}
