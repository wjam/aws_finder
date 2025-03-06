package main

import (
	"context"
	"iter"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi/types"
	"github.com/spf13/cobra"
	"github.com/wjam/aws_finder/internal/finder"
	"github.com/wjam/aws_finder/internal/log"
)

func tagCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tag [key] <value> <value>",
		Short: "Find resources by tags",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return finder.SearchPerRegion(
				cmd.Context(),
				func(ctx context.Context, conf aws.Config) error {
					return findByTag(ctx, resourcegroupstaggingapi.NewFromConfig(conf), args[0], args[1:]...)
				})
		},
	}
}

func findByTag(
	ctx context.Context,
	client resourcegroupstaggingapi.GetResourcesAPIClient,
	key string,
	values ...string,
) error {
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
			return err
		}

		log.Logger(ctx).InfoContext(ctx, aws.ToString(resource.ResourceARN))
	}

	return nil
}

func tagMappingListToResource(r *resourcegroupstaggingapi.GetResourcesOutput) iter.Seq[types.ResourceTagMapping] {
	return slices.Values(r.ResourceTagMappingList)
}
