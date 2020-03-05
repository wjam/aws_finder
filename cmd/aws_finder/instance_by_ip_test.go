package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws/request"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/stretchr/testify/assert"
)

func TestFindInstanceByIp_PrivateIpv4(t *testing.T) {
	var buf bytes.Buffer
	findInstanceByIp(context.TODO(), "find", log.New(&buf, "", 0), &instances{
		[][]*ec2.Reservation{
			{
				{
					Instances: []*ec2.Instance{
						{
							InstanceId: aws.String("skipped"),
							NetworkInterfaces: []*ec2.InstanceNetworkInterface{
								{
									PrivateIpAddresses: []*ec2.InstancePrivateIpAddress{
										{
											PrivateIpAddress: aws.String("nope"),
										},
									},
								},
							},
						},
						{
							InstanceId: aws.String("one to find"),
							NetworkInterfaces: []*ec2.InstanceNetworkInterface{
								{
									PrivateIpAddresses: []*ec2.InstancePrivateIpAddress{
										{
											PrivateIpAddress: aws.String("nope"),
										},
										{
											PrivateIpAddress: aws.String("find"),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	})

	assert.Equal(t, "one to find\n", buf.String())
}

func TestFindInstanceByIp_PrivateIpv6(t *testing.T) {
	var buf bytes.Buffer
	findInstanceByIp(context.TODO(), "find", log.New(&buf, "", 0), &instances{
		[][]*ec2.Reservation{
			{
				{
					Instances: []*ec2.Instance{
						{
							InstanceId: aws.String("skipped"),
							NetworkInterfaces: []*ec2.InstanceNetworkInterface{
								{
									Ipv6Addresses: []*ec2.InstanceIpv6Address{
										{
											Ipv6Address: aws.String("not this one"),
										},
										{
											Ipv6Address: aws.String("neither this one"),
										},
									},
								},
							},
						},
						{
							InstanceId: aws.String("one to find"),
							NetworkInterfaces: []*ec2.InstanceNetworkInterface{
								{
									Ipv6Addresses: []*ec2.InstanceIpv6Address{
										{
											Ipv6Address: aws.String("skipped"),
										},
										{
											Ipv6Address: aws.String("find"),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	})

	assert.Equal(t, "one to find\n", buf.String())
}

func TestFindInstanceByIp_Public(t *testing.T) {
	var buf bytes.Buffer
	findInstanceByIp(context.TODO(), "find", log.New(&buf, "", 0), &instances{
		[][]*ec2.Reservation{
			{
				{
					Instances: []*ec2.Instance{
						{
							InstanceId:      aws.String("not reported"),
							PublicIpAddress: aws.String("skipped"),
						},
						{
							InstanceId:      aws.String("one to find"),
							PublicIpAddress: aws.String("find"),
						},
					},
				},
			},
		},
	})

	assert.Equal(t, "one to find\n", buf.String())
}

var _ instanceLister = &instances{}

type instances struct {
	reservations [][]*ec2.Reservation
}

func (i *instances) DescribeInstancesPagesWithContext(ctx aws.Context, _ *ec2.DescribeInstancesInput, f func(*ec2.DescribeInstancesOutput, bool) bool, _ ...request.Option) error {
	if ctx == nil {
		return fmt.Errorf("missing context")
	}
	for _, r := range i.reservations {
		if !f(&ec2.DescribeInstancesOutput{
			Reservations: r,
		}, true) {
			return fmt.Errorf("expected to search all instances")
		}
	}

	return nil
}
