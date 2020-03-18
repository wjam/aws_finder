package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/stretchr/testify/assert"
)

func TestFindCloudFrontDistributions(t *testing.T) {
	var tests = []struct {
		distributions [][]*cloudfront.DistributionSummary
		needle        string
		expected      string
	}{
		{
			[][]*cloudfront.DistributionSummary{
				{
					{
						Id:         aws.String("unexpected"),
						DomainName: aws.String("unused"),
						Aliases:    &cloudfront.Aliases{},
						Origins:    &cloudfront.Origins{},
					},
				},
				{
					{
						Id:         aws.String("found"),
						DomainName: aws.String("domain-name"),
					},
				},
			},
			"domain-name",
			"found",
		},
		{
			[][]*cloudfront.DistributionSummary{
				{
					{
						Id: aws.String("unused"),
						Aliases: &cloudfront.Aliases{
							Items: aws.StringSlice([]string{"something", "different"}),
						},
						Origins: &cloudfront.Origins{},
					},
				},
				{
					{
						Id: aws.String("found"),
						Aliases: &cloudfront.Aliases{
							Items: aws.StringSlice([]string{"not-this", "alias"}),
						},
						Origins: &cloudfront.Origins{},
					},
				},
			},
			"alias",
			"found",
		},
		{
			[][]*cloudfront.DistributionSummary{
				{
					{
						Id: aws.String("unused"),
						Origins: &cloudfront.Origins{Items: []*cloudfront.Origin{
							{
								DomainName: aws.String("s3.domain"),
							},
						}},
					},
				},
				{
					{
						Id: aws.String("found"),
						Origins: &cloudfront.Origins{Items: []*cloudfront.Origin{
							{
								DomainName: aws.String("s3.domain"),
							},
							{
								DomainName: aws.String("origin"),
							},
						}},
					},
				},
			},
			"origin",
			"found",
		},
	}

	for _, test := range tests {
		t.Run(test.needle, func(t *testing.T) {
			var buf bytes.Buffer
			findCloudFrontDistributions(context.TODO(), test.needle, log.New(&buf, "", 0), &distributions{test.distributions})
			assert.Equal(t, fmt.Sprintf("%s\n", test.expected), buf.String())
		})
	}
}

var _ cloudFrontLister = &distributions{}

type distributions struct {
	distributions [][]*cloudfront.DistributionSummary
}

func (d *distributions) ListDistributionsPagesWithContext(ctx aws.Context, _ *cloudfront.ListDistributionsInput, f func(*cloudfront.ListDistributionsOutput, bool) bool, _ ...request.Option) error {
	if ctx == nil {
		return fmt.Errorf("missing context")
	}
	for _, dist := range d.distributions {
		if !f(&cloudfront.ListDistributionsOutput{
			DistributionList: &cloudfront.DistributionList{Items: dist},
		}, true) {
			return fmt.Errorf("expected to continue")
		}
	}
	return nil
}
