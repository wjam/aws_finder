package main

import (
	"bytes"
	"context"
	"errors"
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

func TestFindVpc(t *testing.T) {
	var buf bytes.Buffer

	ctx := log.ContextWithLogger(t.Context(), slog.New(log.WithAttrsFromContextHandler{
		Parent:            slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}),
		IgnoredAttributes: []string{"time"},
	}))

	require.NoError(t, findVpc(ctx, "needle", &vpcs{
		data: [][]types.Vpc{
			{
				{
					CidrBlock: aws.String("nope"),
					VpcId:     aws.String("not used"),
				},
				{
					CidrBlock: aws.String("something else"),
					VpcId:     aws.String("not used"),
				},
			},
			{
				{
					CidrBlock: aws.String("still nope"),
					VpcId:     aws.String("not used"),
				},
				{
					CidrBlock: aws.String("needle"),
					VpcId:     aws.String("one to find"),
				},
			},
		},
	}))

	assert.Equal(t, "level=INFO msg=\"one to find\"\n", buf.String())
}

var _ ec2.DescribeVpcsAPIClient = &vpcs{}

type vpcs struct {
	data [][]types.Vpc
}

func (v *vpcs) DescribeVpcs(
	ctx context.Context, input *ec2.DescribeVpcsInput, _ ...func(*ec2.Options),
) (*ec2.DescribeVpcsOutput, error) {
	if ctx == nil {
		return nil, errors.New("missing context")
	}
	if len(input.Filters) != 0 || len(input.VpcIds) != 0 {
		return nil, errors.New("invalid input")
	}

	if len(v.data) == 0 {
		return nil, errors.New("no more values")
	}

	var value []types.Vpc
	value, v.data = v.data[0], v.data[1:]

	var token *string
	if len(v.data) != 0 {
		token = aws.String(strconv.Itoa(len(v.data)))
	}

	return &ec2.DescribeVpcsOutput{
		NextToken: token,
		Vpcs:      value,
	}, nil
}
