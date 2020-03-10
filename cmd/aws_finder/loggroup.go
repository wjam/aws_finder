package main

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws/request"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/spf13/cobra"
	"github.com/wjam/aws_finder/internal/finder"
)

func init() {
	commands = append(commands, &cobra.Command{
		Use:   "log_group [needle]",
		Short: "Find a CloudWatch log group by name",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			finder.SearchPerRegion(cmd.Context(), func(ctx context.Context, l *log.Logger, sess *session.Session) {
				findLogGroup(ctx, args[0], l, cloudwatchlogs.New(sess))
			})
		},
	})
}

func findLogGroup(ctx context.Context, needle string, l *log.Logger, client logGroupLister) {
	err := client.DescribeLogGroupsPagesWithContext(ctx, &cloudwatchlogs.DescribeLogGroupsInput{}, func(output *cloudwatchlogs.DescribeLogGroupsOutput, _ bool) bool {
		for _, g := range output.LogGroups {
			if strings.Contains(aws.StringValue(g.LogGroupName), needle) {
				l.Println(aws.StringValue(g.LogGroupName))
			}
		}
		return true
	})
	if err != nil {
		l.Printf("Failed to query instances: %s", err)
		return
	}
}

type logGroupLister interface {
	DescribeLogGroupsPagesWithContext(ctx aws.Context, input *cloudwatchlogs.DescribeLogGroupsInput, fn func(*cloudwatchlogs.DescribeLogGroupsOutput, bool) bool, opts ...request.Option) error
}
