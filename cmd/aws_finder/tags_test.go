package main

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wjam/aws_finder/internal/log"
)

func TestFindByTag(t *testing.T) {
	var buf bytes.Buffer

	ctx := log.ContextWithLogger(t.Context(), slog.New(log.WithAttrsFromContextHandler{
		Parent:            slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}),
		IgnoredAttributes: []string{"time"},
	}))

	require.NoError(t, findByTag(ctx, &resourceTagLister{
		t:      t,
		key:    "tag-key",
		values: []string{"value1", "value2"},
		resources: [][]types.ResourceTagMapping{
			{
				{
					ResourceARN: aws.String("expected"),
				},
			},
		},
	}, "tag-key", "value1", "value2"))

	assert.Equal(t, "level=INFO msg=expected\n", buf.String())
}

var _ resourcegroupstaggingapi.GetResourcesAPIClient = &resourceTagLister{}

type resourceTagLister struct {
	t      *testing.T
	key    string
	values []string

	resources [][]types.ResourceTagMapping
}

func (r *resourceTagLister) GetResources(
	ctx context.Context,
	input *resourcegroupstaggingapi.GetResourcesInput,
	_ ...func(*resourcegroupstaggingapi.Options),
) (*resourcegroupstaggingapi.GetResourcesOutput, error) {
	if ctx == nil {
		return nil, errors.New("missing context")
	}
	if aws.ToBool(input.ExcludeCompliantResources) || len(input.ResourceTypeFilters) != 0 {
		return nil, errors.New("unexpected input")
	}
	if !assert.ElementsMatch(r.t, r.values, input.TagFilters[0].Values) {
		return nil, errors.New("invalid values")
	}
	if len(input.TagFilters) != 1 || aws.ToString(input.TagFilters[0].Key) != r.key {
		return nil, errors.New("invalid input")
	}

	if len(r.resources) == 0 {
		return nil, errors.New("no more values")
	}

	var value []types.ResourceTagMapping
	value, r.resources = r.resources[0], r.resources[1:]

	var token *string
	if len(r.resources) != 0 {
		token = aws.String(strconv.Itoa(len(r.resources)))
	}

	return &resourcegroupstaggingapi.GetResourcesOutput{
		PaginationToken:        token,
		ResourceTagMappingList: value,
	}, nil
}
