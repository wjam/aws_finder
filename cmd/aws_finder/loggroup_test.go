package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws/request"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/stretchr/testify/assert"
)

func TestFindLogGroup(t *testing.T) {
	var buf bytes.Buffer
	findLogGroup(context.TODO(), "find", log.New(&buf, "", 0), &logGroups{
		data: [][]*cloudwatchlogs.LogGroup{
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

var _ logGroupLister = &logGroups{}

type logGroups struct {
	data [][]*cloudwatchlogs.LogGroup
}

func (l *logGroups) DescribeLogGroupsPagesWithContext(ctx aws.Context, input *cloudwatchlogs.DescribeLogGroupsInput, fn func(*cloudwatchlogs.DescribeLogGroupsOutput, bool) bool, _ ...request.Option) error {
	if ctx == nil {
		return fmt.Errorf("missing context")
	}
	if aws.StringValue(input.LogGroupNamePrefix) != "" {
		return fmt.Errorf("invalid prefix")
	}

	for _, page := range l.data {
		if !fn(&cloudwatchlogs.DescribeLogGroupsOutput{LogGroups: page}, true) {
			return fmt.Errorf("should always return true")
		}
	}
	return nil
}
