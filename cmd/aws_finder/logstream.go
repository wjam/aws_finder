package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/go-multierror"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/spf13/cobra"
	"github.com/wjam/aws_finder/internal/finder"
)

func init() {
	commands = append(commands, &cobra.Command{
		Use:   "log_stream [needle]",
		Short: "Find a CloudWatch log stream by name",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			finder.SearchPerRegion(cmd.Context(), func(ctx context.Context, l *log.Logger, sess *session.Session) {
				findLogStream(ctx, args[0], l, cloudwatchlogs.New(sess))
			})
		},
	})
}

func findLogStream(ctx context.Context, needle string, l *log.Logger, client logStreamLister) {
	var errs error

	err := client.DescribeLogGroupsPagesWithContext(ctx, &cloudwatchlogs.DescribeLogGroupsInput{}, func(output *cloudwatchlogs.DescribeLogGroupsOutput, _ bool) bool {
		for _, g := range output.LogGroups {
			if err := findStream(ctx, needle, l, client, aws.StringValue(g.LogGroupName)); err != nil {
				err = multierror.Append(errs, err)
			}
		}
		return true
	})
	if err != nil {
		l.Printf("Failed to query log groups: %s", err)
		return
	}

	if errs != nil {
		l.Printf("Failed to query log streams: %s", err)
		return
	}
}

func findStream(ctx context.Context, needle string, l *log.Logger, client logStreamLister, group string) error {
	return client.DescribeLogStreamsPagesWithContext(ctx, &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName: aws.String(group),
	}, func(output *cloudwatchlogs.DescribeLogStreamsOutput, more bool) bool {
		for _, s := range output.LogStreams {
			if strings.Contains(aws.StringValue(s.LogStreamName), needle) {
				l.Println(fmt.Sprintf("%s/%s", group, aws.StringValue(s.LogStreamName)))
			}
		}
		return true
	})
}

type logStreamLister interface {
	DescribeLogGroupsPagesWithContext(ctx aws.Context, input *cloudwatchlogs.DescribeLogGroupsInput, fn func(*cloudwatchlogs.DescribeLogGroupsOutput, bool) bool, opts ...request.Option) error
	DescribeLogStreamsPagesWithContext(aws.Context, *cloudwatchlogs.DescribeLogStreamsInput, func(*cloudwatchlogs.DescribeLogStreamsOutput, bool) bool, ...request.Option) error
}
