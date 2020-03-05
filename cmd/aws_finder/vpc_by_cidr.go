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
		Use:   "vpc_by_cidr [cidr address]",
		Short: "Find a VPC with the given CIDR range",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			finder.SearchPerRegion(cmd.Context(), func(ctx context.Context, l *log.Logger, sess *session.Session) {
				findVpcByCidr(ctx, args[0], l, ec2.New(sess))
			})
		},
	})
}

func findVpcByCidr(ctx context.Context, needle string, l *log.Logger, client vpcLister) {
	err := client.DescribeVpcsPagesWithContext(ctx, &ec2.DescribeVpcsInput{}, func(output *ec2.DescribeVpcsOutput, _ bool) bool {
		for _, vpc := range output.Vpcs {
			if *vpc.CidrBlock == needle {
				l.Println(*vpc.VpcId)
			}
		}

		return true
	})
	if err != nil {
		l.Printf("Failed to query vpcs: %s", err)
		return
	}
}

type vpcLister interface {
	DescribeVpcsPagesWithContext(ctx aws.Context, input *ec2.DescribeVpcsInput, fn func(*ec2.DescribeVpcsOutput, bool) bool, opts ...request.Option) error
}
