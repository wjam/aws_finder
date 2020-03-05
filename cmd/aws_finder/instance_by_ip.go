package main

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws/request"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/cobra"
	"github.com/wjam/aws_finder/internal/finder"
)

func init() {
	commands = append(commands, &cobra.Command{
		Use:   "instance_by_ip [IP address]",
		Short: "Find an instance with the given IP address",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			finder.SearchPerRegion(cmd.Context(), func(ctx context.Context, l *log.Logger, sess *session.Session) {
				findInstanceByIp(ctx, args[0], l, ec2.New(sess))
			})
		},
	})
}

func findInstanceByIp(ctx context.Context, needle string, l *log.Logger, client instanceLister) {
	err := client.DescribeInstancesPagesWithContext(ctx, &ec2.DescribeInstancesInput{}, func(output *ec2.DescribeInstancesOutput, _ bool) bool {
		for _, r := range output.Reservations {
			for _, instance := range r.Instances {
				if aws.StringValue(instance.PublicIpAddress) == needle {
					l.Println(aws.StringValue(instance.InstanceId))
				}

				for _, network := range instance.NetworkInterfaces {
					for _, ip := range network.PrivateIpAddresses {
						if aws.StringValue(ip.PrivateIpAddress) == needle {
							l.Println(aws.StringValue(instance.InstanceId))
						}
					}
					for _, ip := range network.Ipv6Addresses {
						if aws.StringValue(ip.Ipv6Address) == needle {
							l.Println(aws.StringValue(instance.InstanceId))
						}
					}
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

type instanceLister interface {
	DescribeInstancesPagesWithContext(ctx aws.Context, input *ec2.DescribeInstancesInput, fn func(*ec2.DescribeInstancesOutput, bool) bool, opts ...request.Option) error
}
