package main

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/spf13/cobra"
	"github.com/wjam/aws_finder/internal/finder"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
)

func init() {
	commands = append(commands, &cobra.Command{
		Use:   "s3_by_name [prefix]",
		Short: "Find an S3 bucket by name",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			finder.SearchPerProfile(cmd.Context(), func(ctx context.Context, l *log.Logger, sess *session.Session) {
				findS3ByName(ctx, args[0], l, s3.New(sess))
			})
		},
	})
}

func findS3ByName(ctx context.Context, needle string, l *log.Logger, client s3Lister) {
	buckets, err := client.ListBucketsWithContext(ctx, &s3.ListBucketsInput{})
	if err != nil {
		l.Printf("Failed to query buckets: %s", err)
		return
	}

	for _, bucket := range buckets.Buckets {
		if strings.HasPrefix(aws.StringValue(bucket.Name), needle) {

			location, err := client.GetBucketLocationWithContext(ctx, &s3.GetBucketLocationInput{Bucket: bucket.Name})
			if err != nil {
				l.Printf("Failed to query for bucket '%s' location: %s", aws.StringValue(bucket.Name), err)
				continue
			}

			region := *location.LocationConstraint
			if region == s3.BucketLocationConstraintEu {
				region = s3.BucketLocationConstraintEuWest1
			}
			l.Printf("[%s] %s", region, aws.StringValue(bucket.Name))

		}
	}
}

type s3Lister interface {
	ListBucketsWithContext(ctx aws.Context, input *s3.ListBucketsInput, opts ...request.Option) (*s3.ListBucketsOutput, error)
	GetBucketLocationWithContext(ctx aws.Context, input *s3.GetBucketLocationInput, opts ...request.Option) (*s3.GetBucketLocationOutput, error)
}
