package main

import (
	"log"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/spf13/cobra"
	"github.com/wjam/aws_finder/internal/finder"
)

var vpcByCidr = &cobra.Command{
	Use:   "vpc_by_cidr [cidr address]",
	Short: "Find a VPC with the given CIDR range",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		finder.Search(func(l *log.Logger, sess *session.Session) {
			findVpcByCidr(args[0], l, ec2.New(sess))
		})
	},
}

func findVpcByCidr(needle string, l *log.Logger, client ec2iface.EC2API) {
	err := client.DescribeVpcsPages(&ec2.DescribeVpcsInput{}, func(output *ec2.DescribeVpcsOutput, _ bool) bool {
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
