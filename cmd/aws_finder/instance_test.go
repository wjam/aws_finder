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

func TestFindInstance(t *testing.T) {
	var tests = []struct {
		reservations [][]*ec2.Reservation
		needle       string
		expected     string
	}{
		{
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
								InstanceId: aws.String("found"),
								NetworkInterfaces: []*ec2.InstanceNetworkInterface{
									{
										PrivateIpAddresses: []*ec2.InstancePrivateIpAddress{
											{
												PrivateIpAddress: aws.String("nope"),
											},
											{
												PrivateIpAddress: aws.String("private-ip"),
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"private-ip",
			"found",
		},
		{
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
								InstanceId: aws.String("found"),
								NetworkInterfaces: []*ec2.InstanceNetworkInterface{
									{
										Ipv6Addresses: []*ec2.InstanceIpv6Address{
											{
												Ipv6Address: aws.String("skipped"),
											},
											{
												Ipv6Address: aws.String("ipv6"),
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"ipv6",
			"found",
		},
		{
			[][]*ec2.Reservation{
				{
					{
						Instances: []*ec2.Instance{
							{
								InstanceId: aws.String("not reported"),
								NetworkInterfaces: []*ec2.InstanceNetworkInterface{
									{
										Association: &ec2.InstanceNetworkInterfaceAssociation{
											PublicIp: aws.String("skipped"),
										},
									},
								},
							},
							{
								InstanceId: aws.String("found"),
								NetworkInterfaces: []*ec2.InstanceNetworkInterface{
									{
										Association: &ec2.InstanceNetworkInterfaceAssociation{
											PublicIp: aws.String("public-ip"),
										},
									},
								},
							},
						},
					},
				},
			},
			"public-ip",
			"found",
		},
		{
			[][]*ec2.Reservation{
				{
					{
						Instances: []*ec2.Instance{
							{
								InstanceId:   aws.String("not reported"),
								InstanceType: aws.String("something different"),
							},
							{
								InstanceId:   aws.String("found"),
								InstanceType: aws.String("instance-type"),
							},
						},
					},
				},
			},
			"instance-type",
			"found",
		},
		{
			[][]*ec2.Reservation{
				{
					{
						Instances: []*ec2.Instance{
							{
								InstanceId: aws.String("not reported"),
								ImageId:    aws.String("a public image"),
							},
							{
								InstanceId: aws.String("found"),
								ImageId:    aws.String("image-id"),
							},
						},
					},
				},
			},
			"image-id",
			"found",
		},
	}

	for _, test := range tests {
		t.Run(test.needle, func(t *testing.T) {
			var buf bytes.Buffer
			findInstances(context.TODO(), test.needle, log.New(&buf, "", 0), &instances{test.reservations})
			assert.Equal(t, fmt.Sprintf("%s\n", test.expected), buf.String())
		})
	}
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
