package main

import (
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/wjam/aws_finder/internal/finder"
)

func main() {
	needle := os.Args[1]
	finder.Search(func(l *log.Logger, sess *session.Session) {
		search(needle, l, ec2.New(sess))
	})
}

func search(needle string, l *log.Logger, client ec2iface.EC2API) {
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
