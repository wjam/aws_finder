package main

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wjam/aws_finder/internal/log"
)

func TestFindLogGroup(t *testing.T) {
	var buf bytes.Buffer

	ctx := log.ContextWithLogger(t.Context(), slog.New(log.WithAttrsFromContextHandler{
		Parent:            slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}),
		IgnoredAttributes: []string{"time"},
	}))

	require.NoError(t, findLogGroup(ctx, "find", &logGroups{
		data: [][]types.LogGroup{
			{
				{
					LogGroupName: aws.String("foo"),
				},
				{
					LogGroupName: aws.String("bar"),
				},
			},
			{
				{
					LogGroupName: aws.String("baz"),
				},
				{
					LogGroupName: aws.String("one to find"),
				},
			},
		},
	}))

	assert.Equal(t, `level=INFO msg="one to find"
`, buf.String())
}

var _ cloudwatchlogs.DescribeLogGroupsAPIClient = &logGroups{}

type logGroups struct {
	data [][]types.LogGroup
}

func (l *logGroups) DescribeLogGroups(
	ctx context.Context, input *cloudwatchlogs.DescribeLogGroupsInput, _ ...func(*cloudwatchlogs.Options),
) (*cloudwatchlogs.DescribeLogGroupsOutput, error) {
	if ctx == nil {
		return nil, errors.New("missing context")
	}
	if aws.ToString(input.LogGroupNamePrefix) != "" {
		return nil, errors.New("invalid prefix")
	}

	var value []types.LogGroup
	value, l.data = l.data[0], l.data[1:]

	var token *string
	if len(l.data) != 0 {
		token = aws.String(strconv.Itoa(len(l.data)))
	}

	return &cloudwatchlogs.DescribeLogGroupsOutput{
		LogGroups: value,
		NextToken: token,
	}, nil
}
