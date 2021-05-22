package main

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"

	"github.com/spf13/cobra"
	"github.com/wjam/aws_finder/internal/finder"
)

func init() {
	commands = append(commands, &cobra.Command{
		Use:   "log_group [needle]",
		Short: "Find a CloudWatch log group by name",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			finder.SearchPerRegion(cmd.Context(), func(ctx context.Context, l *log.Logger, conf aws.Config) {
				findLogGroup(ctx, args[0], l, cloudwatchlogs.NewFromConfig(conf))
			})
		},
	})
}

func findLogGroup(ctx context.Context, needle string, l *log.Logger, client cloudwatchlogs.DescribeLogGroupsAPIClient) {
	pages := cloudwatchlogs.NewDescribeLogGroupsPaginator(client, &cloudwatchlogs.DescribeLogGroupsInput{})

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			l.Printf("Failed to query instances: %s", err)
			return
		}

		for _, g := range page.LogGroups {
			if strings.Contains(aws.ToString(g.LogGroupName), needle) {
				l.Println(aws.ToString(g.LogGroupName))
			}
		}
	}
}
