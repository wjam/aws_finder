package main

import (
	"context"
	"fmt"
	"iter"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/spf13/cobra"
	"github.com/wjam/aws_finder/internal/finder"
	"github.com/wjam/aws_finder/internal/log"
)

func logStreamCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "log_stream <logGroupPrefix> [needle]",
		Short: "Find a CloudWatch log stream by name",
		Args:  cobra.RangeArgs(1, 2), //nolint:mnd // up to 2 arguments
		RunE: func(cmd *cobra.Command, args []string) error {
			return finder.SearchPerRegion(
				cmd.Context(),
				func(ctx context.Context, conf aws.Config) error {
					var group *string
					var needle string
					if len(args) == 1 {
						needle = args[0]
					} else {
						group = aws.String(args[0])
						needle = args[1]
					}
					return findLogStream(ctx, group, needle, cloudwatchlogs.NewFromConfig(conf))
				})
		},
	}
}

func findLogStream(
	ctx context.Context, groupPrefix *string, needle string, client logStreamLister,
) error {
	pages := cloudwatchlogs.NewDescribeLogGroupsPaginator(client, &cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePrefix: groupPrefix,
	})

	for g, err := range paginatorToSeq(ctx, pages, logGroupsToLogGroup) {
		if err != nil {
			return err
		}

		if err := findStream(ctx, needle, client, aws.ToString(g.LogGroupName)); err != nil {
			return err
		}
	}

	return nil
}

func logGroupsToLogGroup(r *cloudwatchlogs.DescribeLogGroupsOutput) iter.Seq[types.LogGroup] {
	return slices.Values(r.LogGroups)
}

func findStream(
	ctx context.Context, needle string, client cloudwatchlogs.DescribeLogStreamsAPIClient, group string,
) error {
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

		log.Logger(ctx).InfoContext(ctx, fmt.Sprintf("%s/%s", group, aws.ToString(s.LogStreamName)))
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
