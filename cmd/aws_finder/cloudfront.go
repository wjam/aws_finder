package main

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/spf13/cobra"
	"github.com/wjam/aws_finder/internal/finder"
)

func init() {
	commands = append(commands, &cobra.Command{
		Use:   "cloudfront [needle]",
		Short: "Find CloudFront distributions by domain",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return finder.SearchPerProfile(cmd.Context(), func(ctx context.Context, l *log.Logger, conf aws.Config) error {
				return findCloudFrontDistributions(ctx, args[0], l, cloudfront.NewFromConfig(conf))
			})
		},
	})
}

func findCloudFrontDistributions(ctx context.Context, needle string, l *log.Logger, client cloudfront.ListDistributionsAPIClient) error {
	pages := cloudfront.NewListDistributionsPaginator(client, nil)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return logError("failed to query distributions", err, l)
		}

		for _, dist := range page.DistributionList.Items {
			if findCloudFrontDistribution(needle, dist) {
				l.Println(aws.ToString(dist.Id))
			}
		}
	}

	return nil
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
