package main

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"
	"github.com/spf13/cobra"
	"github.com/wjam/aws_finder/internal/finder"
)

func init() {
	commands = append(commands, &cobra.Command{
		Use:   "tag [key] <value> <value>",
		Short: "Find resources by tags",
		Args:  cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			finder.SearchPerRegion(cmd.Context(), func(ctx context.Context, l *log.Logger, sess *session.Session) {
				findByTag(ctx, resourcegroupstaggingapi.New(sess), l, args[0], args[1:]...)
			})
		},
	})
}

func findByTag(ctx context.Context, client tagPagination, l *log.Logger, key string, values ...string) {
	// TODO need to identify what type of resources the resourcegroupstaggingapi doesn't support
	err := client.GetResourcesPagesWithContext(ctx, &resourcegroupstaggingapi.GetResourcesInput{
		TagFilters: []*resourcegroupstaggingapi.TagFilter{
			{
				Key:    aws.String(key),
				Values: aws.StringSlice(values),
			},
		},
	}, func(output *resourcegroupstaggingapi.GetResourcesOutput, _ bool) bool {
		for _, resource := range output.ResourceTagMappingList {
			l.Println(aws.StringValue(resource.ResourceARN))
		}

		return true
	})
	if err != nil {
		l.Printf("Failed to query tags: %s", err)
		return
	}
}

type tagPagination interface {
	GetResourcesPagesWithContext(aws.Context, *resourcegroupstaggingapi.GetResourcesInput, func(*resourcegroupstaggingapi.GetResourcesOutput, bool) bool, ...request.Option) error
}
