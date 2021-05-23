package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	cloudwatchlogs2 "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindLogStream_AllLogGroups(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, findLogStream(context.Background(), nil, "find", log.New(&buf, "", 0), &logStreams{
		logs: map[string][]types.LogStream{
			"first": {
				{
					LogStreamName: aws.String("skipped"),
				},
				{
					LogStreamName: aws.String("missed"),
				},
				{
					LogStreamName: aws.String("another"),
				},
			},
			"second": {
				{
					LogStreamName: aws.String("not used"),
				},
				{
					LogStreamName: aws.String("one to find"),
				},
			},
			"third": {
				{
					LogStreamName: aws.String("one to miss"),
				},
			},
		},
	}))

	assert.Equal(t, "second/one to find\n", buf.String())
}

func TestFindLogStream_SpecificLogGroups(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, findLogStream(context.Background(), aws.String("expected-prefix"), "find", log.New(&buf, "", 0), &logStreams{
		logStreamPrefix: "expected-prefix",
		logs: map[string][]types.LogStream{
			"expected-prefix": {
				{
					LogStreamName: aws.String("not used"),
				},
				{
					LogStreamName: aws.String("one to find"),
				},
			},
		},
	}))

	assert.Equal(t, "expected-prefix/one to find\n", buf.String())
}

var _ logStreamLister = &logStreams{}

type logStreams struct {
	logStreamPrefix string
	logs            map[string][]types.LogStream
}

func (l *logStreams) DescribeLogGroups(_ context.Context, input *cloudwatchlogs2.DescribeLogGroupsInput, _ ...func(*cloudwatchlogs2.Options)) (*cloudwatchlogs2.DescribeLogGroupsOutput, error) {
	if l.logStreamPrefix != "" {
		if aws.ToString(input.LogGroupNamePrefix) != l.logStreamPrefix {
			return nil, fmt.Errorf("unexpected loggroupnameprefix: %s - %s", l.logStreamPrefix, aws.ToString(input.LogGroupNamePrefix))
		}
	} else if input.LogGroupNamePrefix != nil {
		return nil, fmt.Errorf("unexpected loggroupnameprefix")
	}

	var groups []types.LogGroup
	for name := range l.logs {
		groups = append(groups, types.LogGroup{LogGroupName: aws.String(name)})
	}

	return &cloudwatchlogs2.DescribeLogGroupsOutput{
		LogGroups: groups,
	}, nil
}

func (l *logStreams) DescribeLogStreams(_ context.Context, input *cloudwatchlogs2.DescribeLogStreamsInput, _ ...func(*cloudwatchlogs2.Options)) (*cloudwatchlogs2.DescribeLogStreamsOutput, error) {
	streams := l.logs[aws.ToString(input.LogGroupName)]

	return &cloudwatchlogs2.DescribeLogStreamsOutput{LogStreams: streams}, nil
}
