package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/spf13/cobra"
	"github.com/wjam/aws_finder/internal/finder"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func init() {
	commands = append(commands, &cobra.Command{
		Use:   "s3_bucket [needle]",
		Short: "Find an S3 bucket by name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return finder.SearchPerProfile(cmd.Context(), cmd.OutOrStdout(), func(ctx context.Context, l *log.Logger, conf aws.Config) error {
				return findS3Bucket(ctx, args[0], l, s3.NewFromConfig(conf))
			})
		},
	})
}

func findS3Bucket(ctx context.Context, needle string, l *log.Logger, client s3Lister) error {
	buckets, err := client.ListBuckets(ctx, nil)
	if err != nil {
		return logError("failed to query buckets", err, l)
	}

	for _, bucket := range buckets.Buckets {
		if strings.Contains(aws.ToString(bucket.Name), needle) {

			location, err := client.GetBucketLocation(ctx, &s3.GetBucketLocationInput{Bucket: bucket.Name})
			if err != nil {
				return logError(fmt.Sprintf("failed to query for bucket '%s' location", aws.ToString(bucket.Name)), err, l)
			}

			if location.LocationConstraint == types.BucketLocationConstraintEu {
				location.LocationConstraint = types.BucketLocationConstraintEuWest1
			}
			l.Printf("[%s] %s", location.LocationConstraint, aws.ToString(bucket.Name))

		}
	}

	return nil
}

type s3Lister interface {
	ListBuckets(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error)
	GetBucketLocation(ctx context.Context, params *s3.GetBucketLocationInput, optFns ...func(*s3.Options)) (*s3.GetBucketLocationOutput, error)
}
