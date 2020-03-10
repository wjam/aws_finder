package main

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/cobra"
	"github.com/wjam/aws_finder/internal/finder"
)

func init() {
	commands = append(commands, &cobra.Command{
		Use:   "vpc [needle]",
		Short: "Find a VPC with the given CIDR range",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			finder.SearchPerRegion(cmd.Context(), func(ctx context.Context, l *log.Logger, sess *session.Session) {
				findVpc(ctx, args[0], l, ec2.New(sess))
			})
		},
	})
}

func findVpc(ctx context.Context, needle string, l *log.Logger, client vpcLister) {
	err := client.DescribeVpcsPagesWithContext(ctx, &ec2.DescribeVpcsInput{}, func(output *ec2.DescribeVpcsOutput, _ bool) bool {
		for _, vpc := range output.Vpcs {
			if strings.Contains(*vpc.CidrBlock, needle) {
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
