package main

import (
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/wjam/aws_finder/internal/finder"
)

func main() {
	needle := os.Args[1]
	finder.Search(func(l *log.Logger, sess *session.Session) {
		search(needle, l, ec2.New(sess))
	})
}

func search(needle string, l *log.Logger, client vpcEndpointLister) {
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
