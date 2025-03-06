package main

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/spf13/cobra"
	"github.com/wjam/aws_finder/internal/finder"
	"github.com/wjam/aws_finder/internal/log"
)

func s3BucketCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "s3_bucket [needle]",
		Short: "Find an S3 bucket by name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return finder.SearchPerProfile(
				cmd.Context(),
				func(ctx context.Context, conf aws.Config) error {
					return findS3Bucket(ctx, args[0], s3.NewFromConfig(conf))
				})
		},
	}
}

func findS3Bucket(ctx context.Context, needle string, client s3Lister) error {
	buckets, err := client.ListBuckets(ctx, nil)
	if err != nil {
		return err
	}

	for _, bucket := range buckets.Buckets {
		if strings.Contains(aws.ToString(bucket.Name), needle) {
			location, err := client.GetBucketLocation(ctx, &s3.GetBucketLocationInput{Bucket: bucket.Name})
			if err != nil {
				return fmt.Errorf("failed to query bucket %q for location: %w", aws.ToString(bucket.Name), err)
			}

			if location.LocationConstraint == types.BucketLocationConstraintEu {
				location.LocationConstraint = types.BucketLocationConstraintEuWest1
			}
			log.Logger(ctx).
				InfoContext(
					ctx,
					aws.ToString(bucket.Name),
					slog.String("location", string(location.LocationConstraint)),
				)
		}
	}

	return nil
}

type s3Lister interface {
	ListBuckets(
		ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options),
	) (*s3.ListBucketsOutput, error)
	GetBucketLocation(
		ctx context.Context, params *s3.GetBucketLocationInput, optFns ...func(*s3.Options),
	) (*s3.GetBucketLocationOutput, error)
}
