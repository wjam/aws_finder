package main

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wjam/aws_finder/internal/log"
)

func TestFindS3Bucket(t *testing.T) {
	var buf bytes.Buffer

	ctx := log.ContextWithLogger(t.Context(), slog.New(log.WithAttrsFromContextHandler{
		Parent:            slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}),
		IgnoredAttributes: []string{"time"},
	}))

	require.NoError(t, findS3Bucket(ctx, "find", &buckets{
		buckets: []types.Bucket{
			{
				Name: aws.String("foo"),
			},
			{
				Name: aws.String("bar"),
			},
			{
				Name: aws.String("find me"),
			},
		},
		bucketLocation: map[string]types.BucketLocationConstraint{
			"foo":     "bip",
			"bar":     "baz",
			"find me": "found",
		},
	}))

	assert.Equal(t, "level=INFO msg=\"find me\" location=found\n", buf.String())
}

var _ s3Lister = &buckets{}

type buckets struct {
	buckets        []types.Bucket
	bucketLocation map[string]types.BucketLocationConstraint
}

func (b *buckets) ListBuckets(
	ctx context.Context, _ *s3.ListBucketsInput, _ ...func(*s3.Options),
) (*s3.ListBucketsOutput, error) {
	if ctx == nil {
		return nil, errors.New("missing context")
	}

	return &s3.ListBucketsOutput{
		Buckets: b.buckets,
	}, nil
}

func (b *buckets) GetBucketLocation(
	ctx context.Context, params *s3.GetBucketLocationInput, _ ...func(*s3.Options),
) (*s3.GetBucketLocationOutput, error) {
	if ctx == nil {
		return nil, errors.New("missing context")
	}
	if loc, ok := b.bucketLocation[aws.ToString(params.Bucket)]; ok {
		return &s3.GetBucketLocationOutput{LocationConstraint: loc}, nil
	}

	return nil, errors.New("invalid input")
}
