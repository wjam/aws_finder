package main

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/spf13/cobra"
	"github.com/wjam/aws_finder/internal/finder"
)

func init() {
	commands = append(commands, &cobra.Command{
		Use:   "instance [needle]",
		Short: "Find an instance by type, AMI or ip address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return finder.SearchPerRegion(cmd.Context(), func(ctx context.Context, l *log.Logger, conf aws.Config) error {
				return findInstances(ctx, args[0], l, ec2.NewFromConfig(conf))
			})
		},
	})
}

func findInstances(ctx context.Context, needle string, l *log.Logger, client ec2.DescribeInstancesAPIClient) error {
	pages := ec2.NewDescribeInstancesPaginator(client, nil)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return logError("failed to query instances", err, l)
		}

		for _, r := range page.Reservations {
			for _, instance := range r.Instances {
				if findInstance(needle, instance) {
					l.Println(aws.ToString(instance.InstanceId))
				}
			}
		}
	}

	return nil
}

func findInstance(needle string, instance types.Instance) bool {
	if check(needle, instance.ImageId, aws.String(string(instance.InstanceType))) {
		return true
	}

	for _, network := range instance.NetworkInterfaces {
		for _, ip := range network.PrivateIpAddresses {
			if check(needle, ip.PrivateIpAddress) {
				return true
			}
		}
		for _, ip := range network.Ipv6Addresses {
			if check(needle, ip.Ipv6Address) {
				return true
			}
		}
		if network.Association != nil && check(needle, network.Association.PublicIp) {
			return true
		}
	}

	return false
}

func check(needle string, haystack ...*string) bool {
	for _, item := range haystack {
		if strings.Contains(aws.ToString(item), needle) {
			return true
		}
	}
	return false
}
