package main

import (
	"context"
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
		Run: func(cmd *cobra.Command, args []string) {
			finder.SearchPerProfile(cmd.Context(), func(ctx context.Context, l *log.Logger, conf aws.Config) {
				findS3Bucket(ctx, args[0], l, s3.NewFromConfig(conf))
			})
		},
	})
}

func findS3Bucket(ctx context.Context, needle string, l *log.Logger, client s3Lister) {
	buckets, err := client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		l.Printf("Failed to query buckets: %s", err)
		return
	}

	for _, bucket := range buckets.Buckets {
		if strings.Contains(aws.ToString(bucket.Name), needle) {

			location, err := client.GetBucketLocation(ctx, &s3.GetBucketLocationInput{Bucket: bucket.Name})
			if err != nil {
				l.Printf("Failed to query for bucket '%s' location: %s", aws.ToString(bucket.Name), err)
				continue
			}

			if location.LocationConstraint == types.BucketLocationConstraintEu {
				location.LocationConstraint = types.BucketLocationConstraintEuWest1
			}
			l.Printf("[%s] %s", location.LocationConstraint, aws.ToString(bucket.Name))

		}
	}
}

type s3Lister interface {
	ListBuckets(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error)
	GetBucketLocation(ctx context.Context, params *s3.GetBucketLocationInput, optFns ...func(*s3.Options)) (*s3.GetBucketLocationOutput, error)
}
