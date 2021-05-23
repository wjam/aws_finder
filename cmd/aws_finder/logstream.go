package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/spf13/cobra"
	"github.com/wjam/aws_finder/internal/finder"
)

func init() {
	commands = append(commands, &cobra.Command{
		Use:   "log_stream <logGroupPrefix> [needle]",
		Short: "Find a CloudWatch log stream by name",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return finder.SearchPerRegion(cmd.Context(), func(ctx context.Context, l *log.Logger, conf aws.Config) error {
				var group *string
				var needle string
				if len(args) == 1 {
					needle = args[0]
				} else {
					group = aws.String(args[0])
					needle = args[1]
				}
				return findLogStream(ctx, group, needle, l, cloudwatchlogs.NewFromConfig(conf))
			})
		},
	})
}

func findLogStream(ctx context.Context, groupPrefix *string, needle string, l *log.Logger, client logStreamLister) error {
	pages := cloudwatchlogs.NewDescribeLogGroupsPaginator(client, &cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePrefix: groupPrefix,
	})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return logError("failed to query log groups", err, l)
		}

		for _, g := range page.LogGroups {
			if err := findStream(ctx, needle, l, client, aws.ToString(g.LogGroupName)); err != nil {
				return logError("failed to query log streams", err, l)
			}
		}
	}

	return nil
}

func findStream(ctx context.Context, needle string, l *log.Logger, client cloudwatchlogs.DescribeLogStreamsAPIClient, group string) error {
	pages := cloudwatchlogs.NewDescribeLogStreamsPaginator(client, &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName: aws.String(group),
	})

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return err
		}

		for _, s := range page.LogStreams {
			if strings.Contains(aws.ToString(s.LogStreamName), needle) {
				l.Println(fmt.Sprintf("%s/%s", group, aws.ToString(s.LogStreamName)))
			}
		}
	}

	return nil
}

type logStreamLister interface {
	cloudwatchlogs.DescribeLogGroupsAPIClient
	cloudwatchlogs.DescribeLogStreamsAPIClient
}
