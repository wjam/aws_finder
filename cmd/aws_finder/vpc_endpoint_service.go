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
		RunE: func(cmd *cobra.Command, args []string) error {
			return finder.SearchPerRegion(cmd.Context(), func(ctx context.Context, l *log.Logger, conf aws.Config) error {
				return findVpcEndpointService(ctx, args[0], l, ec2.NewFromConfig(conf))
			})
		},
	})
}

func findVpcEndpointService(ctx context.Context, needle string, l *log.Logger, client describeVpcEndpointServicesClient) error {
	pages := newDescribeVpcEndpointServicesPaginator(client, nil)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return logError("failed to query vpc endpoint services", err, l)
		}

		for _, svc := range page.ServiceDetails {
			if strings.Contains(*svc.ServiceName, needle) {
				l.Printf(*svc.ServiceName)
			}
		}
	}

	return nil
}
