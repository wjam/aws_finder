package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/stretchr/testify/assert"
)

func TestFindLogStream_AllLogGroups(t *testing.T) {
	var buf bytes.Buffer
	findLogStream(context.TODO(), nil, "find", log.New(&buf, "", 0), &logStreams{
		logs: map[string][]*cloudwatchlogs.LogStream{
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
	})

	assert.Equal(t, "second/one to find\n", buf.String())
}

func TestFindLogStream_SpecificLogGroups(t *testing.T) {
	var buf bytes.Buffer
	findLogStream(context.TODO(), aws.String("expected-prefix"), "find", log.New(&buf, "", 0), &logStreams{
		logStreamPrefix: "expected-prefix",
		logs: map[string][]*cloudwatchlogs.LogStream{
			"expected-prefix": {
				{
					LogStreamName: aws.String("not used"),
				},
				{
					LogStreamName: aws.String("one to find"),
				},
			},
		},
	})

	assert.Equal(t, "expected-prefix/one to find\n", buf.String())
}

var _ logStreamLister = &logStreams{}

type logStreams struct {
	logStreamPrefix string
	logs            map[string][]*cloudwatchlogs.LogStream
}

func (l *logStreams) DescribeLogGroupsPagesWithContext(_ aws.Context, input *cloudwatchlogs.DescribeLogGroupsInput, f func(*cloudwatchlogs.DescribeLogGroupsOutput, bool) bool, _ ...request.Option) error {
	if l.logStreamPrefix != "" {
		if aws.StringValue(input.LogGroupNamePrefix) != l.logStreamPrefix {
			return fmt.Errorf("unexpected loggroupnameprefix: %s - %s", l.logStreamPrefix, aws.StringValue(input.LogGroupNamePrefix))
		}
	} else if input.LogGroupNamePrefix != nil {
		return fmt.Errorf("unexpected loggroupnameprefix")
	}

	var groups []*cloudwatchlogs.LogGroup
	for name := range l.logs {
		groups = append(groups, &cloudwatchlogs.LogGroup{LogGroupName: aws.String(name)})
	}

	if !f(&cloudwatchlogs.DescribeLogGroupsOutput{LogGroups: groups}, false) {
		return fmt.Errorf("should always continue")
	}

	return nil
}

func (l *logStreams) DescribeLogStreamsPagesWithContext(_ aws.Context, input *cloudwatchlogs.DescribeLogStreamsInput, f func(*cloudwatchlogs.DescribeLogStreamsOutput, bool) bool, _ ...request.Option) error {
	streams := l.logs[aws.StringValue(input.LogGroupName)]

	if !f(&cloudwatchlogs.DescribeLogStreamsOutput{LogStreams: streams}, false) {
		return fmt.Errorf("should always continue")
	}

	return nil
}
