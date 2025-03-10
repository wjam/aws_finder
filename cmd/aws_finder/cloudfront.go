package main

import (
	"context"
	"iter"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/spf13/cobra"
	"github.com/wjam/aws_finder/internal/finder"
	"github.com/wjam/aws_finder/internal/log"
)

func cloudfrontCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cloudfront [needle]",
		Short: "Find CloudFront distributions by domain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return finder.SearchPerProfile(
				cmd.Context(),
				func(ctx context.Context, conf aws.Config) error {
					return findCloudFrontDistributions(ctx, args[0], cloudfront.NewFromConfig(conf))
				})
		},
	}
}

func findCloudFrontDistributions(
	ctx context.Context, needle string, client cloudfront.ListDistributionsAPIClient,
) error {
	pages := cloudfront.NewListDistributionsPaginator(client, nil)

	seq := paginatorToSeq(ctx, pages, cloudfrontListToItems)
	seq = filter2(func(dist types.DistributionSummary, err error) bool {
		return err != nil || findCloudFrontDistribution(needle, dist)
	}, seq)

	for dist, err := range seq {
		if err != nil {
			return err
		}

		log.Logger(ctx).InfoContext(ctx, aws.ToString(dist.Id))
	}

	return nil
}

func cloudfrontListToItems(r *cloudfront.ListDistributionsOutput) iter.Seq[types.DistributionSummary] {
	return slices.Values(r.DistributionList.Items)
}

func findCloudFrontDistribution(needle string, dist types.DistributionSummary) bool {
	var values = []*string{dist.DomainName}
	if dist.Aliases != nil && dist.Aliases.Items != nil {
		values = append(values, aws.StringSlice(dist.Aliases.Items)...)
	}
	if dist.Origins != nil {
		for _, name := range dist.Origins.Items {
			values = append(values, name.DomainName)
		}
	}
	return check(needle, values...)
}
