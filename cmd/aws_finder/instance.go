package main

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws/request"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/cobra"
	"github.com/wjam/aws_finder/internal/finder"
)

func init() {
	commands = append(commands, &cobra.Command{
		Use:   "instance [needle]",
		Short: "Find an instance by type, AMI or ip address",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			finder.SearchPerRegion(cmd.Context(), func(ctx context.Context, l *log.Logger, sess *session.Session) {
				findInstances(ctx, args[0], l, ec2.New(sess))
			})
		},
	})
}

func findInstances(ctx context.Context, needle string, l *log.Logger, client instanceLister) {
	err := client.DescribeInstancesPagesWithContext(ctx, &ec2.DescribeInstancesInput{}, func(output *ec2.DescribeInstancesOutput, _ bool) bool {
		for _, r := range output.Reservations {
			for _, instance := range r.Instances {
				if findInstance(needle, instance) {
					l.Println(aws.StringValue(instance.InstanceId))
				}
			}
		}
		return true
	})
	if err != nil {
		l.Printf("Failed to query instances: %s", err)
		return
	}
}

func findInstance(needle string, instance *ec2.Instance) bool {
	if check(needle, instance.ImageId, instance.InstanceType) {
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
		if strings.Contains(aws.StringValue(item), needle) {
			return true
		}
	}
	return false
}

type instanceLister interface {
	DescribeInstancesPagesWithContext(ctx aws.Context, input *ec2.DescribeInstancesInput, fn func(*ec2.DescribeInstancesOutput, bool) bool, opts ...request.Option) error
}
