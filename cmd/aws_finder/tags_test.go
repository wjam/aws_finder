package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"
	"github.com/stretchr/testify/assert"
)

func TestFindByTag(t *testing.T) {
	var buf bytes.Buffer
	findByTag(context.TODO(), &resourceTagLister{
		t:      t,
		key:    "tag-key",
		values: []string{"value1", "value2"},
		resources: [][]*resourcegroupstaggingapi.ResourceTagMapping{
			{
				{
					ResourceARN: aws.String("expected"),
				},
			},
		},
	}, log.New(&buf, "", 0), "tag-key", "value1", "value2")

	assert.Equal(t, "expected\n", buf.String())
}

var _ tagPagination = &resourceTagLister{}

type resourceTagLister struct {
	t      *testing.T
	key    string
	values []string

	resources [][]*resourcegroupstaggingapi.ResourceTagMapping
}

func (r *resourceTagLister) GetResourcesPagesWithContext(ctx aws.Context, input *resourcegroupstaggingapi.GetResourcesInput, f func(*resourcegroupstaggingapi.GetResourcesOutput, bool) bool, _ ...request.Option) error {
	if ctx == nil {
		return fmt.Errorf("missing context")
	}
	if aws.BoolValue(input.ExcludeCompliantResources) || len(input.ResourceTypeFilters) != 0 {
		return fmt.Errorf("unexpected input")
	}
	if !assert.ElementsMatch(r.t, r.values, aws.StringValueSlice(input.TagFilters[0].Values)) {
		return fmt.Errorf("invalid values")
	}
	if len(input.TagFilters) != 1 || aws.StringValue(input.TagFilters[0].Key) != r.key {
		return fmt.Errorf("invalid input")
	}

	for _, resources := range r.resources {
		if !f(&resourcegroupstaggingapi.GetResourcesOutput{
			ResourceTagMappingList: resources,
		}, true) {
			return fmt.Errorf("should always continue")
		}
	}

	return nil
}
