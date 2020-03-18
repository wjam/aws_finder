package main

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/spf13/cobra"
	"github.com/wjam/aws_finder/internal/finder"
)

func init() {
	commands = append(commands, &cobra.Command{
		Use:   "cloudfront [needle]",
		Short: "Find CloudFront distributions by domain",
		Args:  cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			finder.SearchPerProfile(cmd.Context(), func(ctx context.Context, l *log.Logger, sess *session.Session) {
				findCloudFrontDistributions(ctx, args[0], l, cloudfront.New(sess))
			})
		},
	})
}

func findCloudFrontDistributions(ctx context.Context, needle string, l *log.Logger, client cloudFrontLister) {
	err := client.ListDistributionsPagesWithContext(ctx, &cloudfront.ListDistributionsInput{}, func(output *cloudfront.ListDistributionsOutput, more bool) bool {
		for _, dist := range output.DistributionList.Items {
			if findCloudFrontDistribution(needle, dist) {
				l.Println(aws.StringValue(dist.Id))
			}
		}
		return true
	})
	if err != nil {
		l.Printf("Failed to query distributions: %s", err)
		return
	}
}

func findCloudFrontDistribution(needle string, dist *cloudfront.DistributionSummary) bool {
	var values = []*string{dist.DomainName}
	if dist.Aliases != nil && dist.Aliases.Items != nil {
		values = append(values, dist.Aliases.Items...)
	}
	if dist.Origins != nil {
		for _, name := range dist.Origins.Items {
			values = append(values, name.DomainName)
		}
	}
	return check(needle, values...)
}

type cloudFrontLister interface {
	ListDistributionsPagesWithContext(aws.Context, *cloudfront.ListDistributionsInput, func(*cloudfront.ListDistributionsOutput, bool) bool, ...request.Option) error
}
