package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindByTag(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, findByTag(context.Background(), &resourceTagLister{
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
	}, log.New(&buf, "", 0), "tag-key", "value1", "value2"))

	assert.Equal(t, "expected\n", buf.String())
}

var _ resourcegroupstaggingapi.GetResourcesAPIClient = &resourceTagLister{}

type resourceTagLister struct {
	t      *testing.T
	key    string
	values []string

	resources [][]types.ResourceTagMapping
}

func (r *resourceTagLister) GetResources(ctx context.Context, input *resourcegroupstaggingapi.GetResourcesInput, _ ...func(*resourcegroupstaggingapi.Options)) (*resourcegroupstaggingapi.GetResourcesOutput, error) {
	if ctx == nil {
		return nil, fmt.Errorf("missing context")
	}
	if aws.ToBool(input.ExcludeCompliantResources) || len(input.ResourceTypeFilters) != 0 {
		return nil, fmt.Errorf("unexpected input")
	}
	if !assert.ElementsMatch(r.t, r.values, input.TagFilters[0].Values) {
		return nil, fmt.Errorf("invalid values")
	}
	if len(input.TagFilters) != 1 || aws.ToString(input.TagFilters[0].Key) != r.key {
		return nil, fmt.Errorf("invalid input")
	}

	if len(r.resources) == 0 {
		return nil, fmt.Errorf("no more values")
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
