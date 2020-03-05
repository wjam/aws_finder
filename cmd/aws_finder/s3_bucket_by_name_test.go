package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/assert"
)

func TestFindS3ByName(t *testing.T) {
	var buf bytes.Buffer
	findS3ByName(context.TODO(), "find", log.New(&buf, "", 0), &buckets{
		buckets: []*s3.Bucket{
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
		bucketLocation: map[string]string{
			"foo":     "bip",
			"bar":     "baz",
			"find me": "found",
		},
	})

	assert.Equal(t, "[found] find me\n", buf.String())
}

type buckets struct {
	buckets        []*s3.Bucket
	bucketLocation map[string]string
}

func (b *buckets) ListBucketsWithContext(ctx aws.Context, _ *s3.ListBucketsInput, _ ...request.Option) (*s3.ListBucketsOutput, error) {
	if ctx == nil {
		return nil, fmt.Errorf("missing context")
	}

	return &s3.ListBucketsOutput{
		Buckets: b.buckets,
	}, nil
}

func (b *buckets) GetBucketLocationWithContext(ctx aws.Context, input *s3.GetBucketLocationInput, _ ...request.Option) (*s3.GetBucketLocationOutput, error) {
	if ctx == nil {
		return nil, fmt.Errorf("missing context")
	}
	if loc, ok := b.bucketLocation[aws.StringValue(input.Bucket)]; ok {
		return &s3.GetBucketLocationOutput{LocationConstraint: aws.String(loc)}, nil
	}

	return nil, fmt.Errorf("invalid input")
}
