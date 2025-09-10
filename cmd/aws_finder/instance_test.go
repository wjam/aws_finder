package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wjam/aws_finder/internal/log"
)

func TestFindInstance(t *testing.T) {
	var tests = []struct {
		reservations [][]types.Reservation
		needle       string
		expected     string
	}{
		{
			[][]types.Reservation{
				{
					{
						Instances: []types.Instance{
							{
								InstanceId: aws.String("skipped"),
								NetworkInterfaces: []types.InstanceNetworkInterface{
									{
										PrivateIpAddresses: []types.InstancePrivateIpAddress{
											{
												PrivateIpAddress: aws.String("nope"),
											},
										},
									},
								},
							},
							{
								InstanceId: aws.String("found"),
								NetworkInterfaces: []types.InstanceNetworkInterface{
									{
										PrivateIpAddresses: []types.InstancePrivateIpAddress{
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
			[][]types.Reservation{
				{
					{
						Instances: []types.Instance{
							{
								InstanceId: aws.String("skipped"),
								NetworkInterfaces: []types.InstanceNetworkInterface{
									{
										Ipv6Addresses: []types.InstanceIpv6Address{
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
								NetworkInterfaces: []types.InstanceNetworkInterface{
									{
										Ipv6Addresses: []types.InstanceIpv6Address{
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
			[][]types.Reservation{
				{
					{
						Instances: []types.Instance{
							{
								InstanceId: aws.String("not reported"),
								NetworkInterfaces: []types.InstanceNetworkInterface{
									{
										Association: &types.InstanceNetworkInterfaceAssociation{
											PublicIp: aws.String("skipped"),
										},
									},
								},
							},
							{
								InstanceId: aws.String("found"),
								NetworkInterfaces: []types.InstanceNetworkInterface{
									{
										Association: &types.InstanceNetworkInterfaceAssociation{
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
			[][]types.Reservation{
				{
					{
						Instances: []types.Instance{
							{
								InstanceId:   aws.String("not reported"),
								InstanceType: "something different",
							},
							{
								InstanceId:   aws.String("found"),
								InstanceType: "instance-type",
							},
						},
					},
				},
			},
			"instance-type",
			"found",
		},
		{
			[][]types.Reservation{
				{
					{
						Instances: []types.Instance{
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

			ctx := log.ContextWithLogger(t.Context(), slog.New(log.WithAttrsFromContextHandler{
				Parent: slog.NewTextHandler(io.MultiWriter(&buf, t.Output()), &slog.HandlerOptions{
					Level:       slog.LevelDebug,
					ReplaceAttr: log.FilterAttributesFromLog([]string{"time"}),
				}),
			}))

			require.NoError(t, findInstances(ctx, test.needle, &instances{test.reservations}))
			assert.Equal(t, fmt.Sprintf("level=INFO msg=%s\n", test.expected), buf.String())
		})
	}
}

var _ ec2.DescribeInstancesAPIClient = &instances{}

type instances struct {
	reservations [][]types.Reservation
}

func (i *instances) DescribeInstances(
	ctx context.Context, _ *ec2.DescribeInstancesInput, _ ...func(*ec2.Options),
) (*ec2.DescribeInstancesOutput, error) {
	if ctx == nil {
		return nil, errors.New("missing context")
	}

	var value []types.Reservation
	value, i.reservations = i.reservations[0], i.reservations[1:]

	var token *string
	if len(i.reservations) != 0 {
		token = aws.String(strconv.Itoa(len(i.reservations)))
	}

	return &ec2.DescribeInstancesOutput{
		NextToken:    token,
		Reservations: value,
	}, nil
}
