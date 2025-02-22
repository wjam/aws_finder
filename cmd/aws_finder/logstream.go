package main

import (
	"context"
	"iter"
	"log"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
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

	for g, err := range paginatorToSeq(ctx, pages, logGroupsToLogGroup) {
		if err != nil {
			return logError("failed to query log groups", err, l)
		}

		if err := findStream(ctx, needle, l, client, aws.ToString(g.LogGroupName)); err != nil {
			return logError("failed to query log streams", err, l)
		}
	}

	return nil
}

func logGroupsToLogGroup(r *cloudwatchlogs.DescribeLogGroupsOutput) iter.Seq[types.LogGroup] {
	return slices.Values(r.LogGroups)
}

func findStream(ctx context.Context, needle string, l *log.Logger, client cloudwatchlogs.DescribeLogStreamsAPIClient, group string) error {
	pages := cloudwatchlogs.NewDescribeLogStreamsPaginator(client, &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName: aws.String(group),
	})

	seq := paginatorToSeq(ctx, pages, logStreamToLogStream)
	seq = filter2(func(s types.LogStream, err error) bool {
		return err != nil || strings.Contains(aws.ToString(s.LogStreamName), needle)
	}, seq)

	for s, err := range seq {
		if err != nil {
			return err
		}

		l.Printf("%s/%s\n", group, aws.ToString(s.LogStreamName))
	}

	return nil
}

func logStreamToLogStream(r *cloudwatchlogs.DescribeLogStreamsOutput) iter.Seq[types.LogStream] {
	return slices.Values(r.LogStreams)
}

type logStreamLister interface {
	cloudwatchlogs.DescribeLogGroupsAPIClient
	cloudwatchlogs.DescribeLogStreamsAPIClient
}
