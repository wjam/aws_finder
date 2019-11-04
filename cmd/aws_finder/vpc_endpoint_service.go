package main

import (
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/cobra"
	"github.com/wjam/aws_finder/internal/finder"
)

var vpcEndpointService = &cobra.Command{
	Use:   "vpc_endpoint_service [service name]",
	Short: "Find a VPC endpoint service by the given service name",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		finder.Search(func(l *log.Logger, sess *session.Session) {
			findVpcEndpointService(args[0], l, ec2.New(sess))
		})
	},
}

func findVpcEndpointService(needle string, l *log.Logger, client vpcEndpointLister) {
	var next *string
	for {
		output, err := client.DescribeVpcEndpointServices(&ec2.DescribeVpcEndpointServicesInput{
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

type vpcEndpointLister interface {
	DescribeVpcEndpointServices(*ec2.DescribeVpcEndpointServicesInput) (*ec2.DescribeVpcEndpointServicesOutput, error)
}
