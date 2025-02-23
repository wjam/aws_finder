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
		Use:   "log_group [needle]",
		Short: "Find a CloudWatch log group by name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return finder.SearchPerRegion(cmd.Context(), cmd.OutOrStdout(), func(ctx context.Context, l *log.Logger, conf aws.Config) error {
				return findLogGroup(ctx, args[0], l, cloudwatchlogs.NewFromConfig(conf))
			})
		},
	})
}

func findLogGroup(ctx context.Context, needle string, l *log.Logger, client cloudwatchlogs.DescribeLogGroupsAPIClient) error {
	pages := cloudwatchlogs.NewDescribeLogGroupsPaginator(client, nil)

	seq := paginatorToSeq(ctx, pages, logGroupListToItems)
	seq = filter2(func(g types.LogGroup, err error) bool {
		return err != nil || strings.Contains(aws.ToString(g.LogGroupName), needle)
	}, seq)

	for g, err := range seq {
		if err != nil {
			return logError("failed to query instances", err, l)
		}

		l.Println(aws.ToString(g.LogGroupName))
	}

	return nil
}

func logGroupListToItems(r *cloudwatchlogs.DescribeLogGroupsOutput) iter.Seq[types.LogGroup] {
	return slices.Values(r.LogGroups)
}
