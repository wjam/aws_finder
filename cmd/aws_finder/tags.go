package main

import (
	"context"
	"iter"
	"log"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi/types"
	"github.com/spf13/cobra"
	"github.com/wjam/aws_finder/internal/finder"
)

func init() {
	commands = append(commands, &cobra.Command{
		Use:   "tag [key] <value> <value>",
		Short: "Find resources by tags",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return finder.SearchPerRegion(cmd.Context(), func(ctx context.Context, l *log.Logger, conf aws.Config) error {
				return findByTag(ctx, resourcegroupstaggingapi.NewFromConfig(conf), l, args[0], args[1:]...)
			})
		},
	})
}

func findByTag(ctx context.Context, client resourcegroupstaggingapi.GetResourcesAPIClient, l *log.Logger, key string, values ...string) error {
	// TODO need to identify what type of resources the resourcegroupstaggingapi doesn't support

	pages := resourcegroupstaggingapi.NewGetResourcesPaginator(client, &resourcegroupstaggingapi.GetResourcesInput{
		TagFilters: []types.TagFilter{
			{
				Key:    aws.String(key),
				Values: values,
			},
		},
	})

	for resource, err := range paginatorToSeq(ctx, pages, tagMappingListToResource) {
		if err != nil {
			return logError("failed to query tags", err, l)
		}

		l.Println(aws.ToString(resource.ResourceARN))
	}

	return nil
}

func tagMappingListToResource(r *resourcegroupstaggingapi.GetResourcesOutput) iter.Seq[types.ResourceTagMapping] {
	return slices.Values(r.ResourceTagMappingList)
}
