package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/stretchr/testify/assert"
)

func TestFindLogGroup(t *testing.T) {
	var buf bytes.Buffer
	findLogGroup(context.TODO(), "find", log.New(&buf, "", 0), &logGroups{
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
	})

	assert.Equal(t, "one to find\n", buf.String())
}

var _ cloudwatchlogs.DescribeLogGroupsAPIClient = &logGroups{}

type logGroups struct {
	data [][]types.LogGroup
}

func (l *logGroups) DescribeLogGroups(ctx context.Context, input *cloudwatchlogs.DescribeLogGroupsInput, _ ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error) {
	if ctx == nil {
		return nil, fmt.Errorf("missing context")
	}
	if aws.ToString(input.LogGroupNamePrefix) != "" {
		return nil, fmt.Errorf("invalid prefix")
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
