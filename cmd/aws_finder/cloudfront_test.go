package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wjam/aws_finder/internal/log"
)

func TestFindCloudFrontDistributions(t *testing.T) {
	var tests = []struct {
		distributions [][]types.DistributionSummary
		needle        string
		expected      string
	}{
		{
			[][]types.DistributionSummary{
				{
					{
						Id:         aws.String("unexpected"),
						DomainName: aws.String("unused"),
						Aliases:    &types.Aliases{},
						Origins:    &types.Origins{},
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
			[][]types.DistributionSummary{
				{
					{
						Id: aws.String("unused"),
						Aliases: &types.Aliases{
							Items: []string{"something", "different"},
						},
						Origins: &types.Origins{},
					},
				},
				{
					{
						Id: aws.String("found"),
						Aliases: &types.Aliases{
							Items: []string{"not-this", "alias"},
						},
						Origins: &types.Origins{},
					},
				},
			},
			"alias",
			"found",
		},
		{
			[][]types.DistributionSummary{
				{
					{
						Id: aws.String("unused"),
						Origins: &types.Origins{Items: []types.Origin{
							{
								DomainName: aws.String("s3.domain"),
							},
						}},
					},
				},
				{
					{
						Id: aws.String("found"),
						Origins: &types.Origins{Items: []types.Origin{
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

			ctx := log.ContextWithLogger(t.Context(), slog.New(log.WithAttrsFromContextHandler{
				Parent:            slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}),
				IgnoredAttributes: []string{"time"},
			}))

			err := findCloudFrontDistributions(ctx, test.needle, &distributions{test.distributions})
			require.NoError(t, err)
			assert.Equal(t, fmt.Sprintf("level=INFO msg=%s\n", test.expected), buf.String())
		})
	}
}

var _ cloudfront.ListDistributionsAPIClient = &distributions{}

type distributions struct {
	distributions [][]types.DistributionSummary
}

func (d *distributions) ListDistributions(
	ctx context.Context, _ *cloudfront.ListDistributionsInput, _ ...func(*cloudfront.Options),
) (*cloudfront.ListDistributionsOutput, error) {
	if ctx == nil {
		return nil, errors.New("missing context")
	}

	var value []types.DistributionSummary
	value, d.distributions = d.distributions[0], d.distributions[1:]

	var token *string
	if len(d.distributions) != 0 {
		token = aws.String(strconv.Itoa(len(d.distributions)))
	}

	return &cloudfront.ListDistributionsOutput{
		DistributionList: &types.DistributionList{
			Items:      value,
			NextMarker: token,
		},
	}, nil
}
