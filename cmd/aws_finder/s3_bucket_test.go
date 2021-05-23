package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindS3Bucket(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, findS3Bucket(context.Background(), "find", log.New(&buf, "", 0), &buckets{
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

	assert.Equal(t, "[found] find me\n", buf.String())
}

var _ s3Lister = &buckets{}

type buckets struct {
	buckets        []types.Bucket
	bucketLocation map[string]types.BucketLocationConstraint
}

func (b *buckets) ListBuckets(ctx context.Context, _ *s3.ListBucketsInput, _ ...func(*s3.Options)) (*s3.ListBucketsOutput, error) {
	if ctx == nil {
		return nil, fmt.Errorf("missing context")
	}

	return &s3.ListBucketsOutput{
		Buckets: b.buckets,
	}, nil
}

func (b *buckets) GetBucketLocation(ctx context.Context, params *s3.GetBucketLocationInput, _ ...func(*s3.Options)) (*s3.GetBucketLocationOutput, error) {
	if ctx == nil {
		return nil, fmt.Errorf("missing context")
	}
	if loc, ok := b.bucketLocation[aws.ToString(params.Bucket)]; ok {
		return &s3.GetBucketLocationOutput{LocationConstraint: loc}, nil
	}

	return nil, fmt.Errorf("invalid input")
}
