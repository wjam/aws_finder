package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws/request"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/stretchr/testify/assert"
)

func TestFindVpc(t *testing.T) {
	var buf bytes.Buffer
	findVpc(context.TODO(), "needle", log.New(&buf, "", 0), &vpcs{
		data: [][]*ec2.Vpc{
			{
				{
					CidrBlock: aws.String("nope"),
					VpcId:     aws.String("not used"),
				},
				{
					CidrBlock: aws.String("something else"),
					VpcId:     aws.String("not used"),
				},
			},
			{
				{
					CidrBlock: aws.String("still nope"),
					VpcId:     aws.String("not used"),
				},
				{
					CidrBlock: aws.String("needle"),
					VpcId:     aws.String("one to find"),
				},
			},
		},
	})

	assert.Equal(t, "one to find\n", buf.String())
}

var _ vpcLister = &vpcs{}

type vpcs struct {
	data [][]*ec2.Vpc
}

func (v *vpcs) DescribeVpcsPagesWithContext(ctx aws.Context, input *ec2.DescribeVpcsInput, fn func(*ec2.DescribeVpcsOutput, bool) bool, _ ...request.Option) error {
	if ctx == nil {
		return fmt.Errorf("missing context")
	}
	if len(input.Filters) != 0 || len(input.VpcIds) != 0 {
		return fmt.Errorf("invalid input")
	}

	for _, page := range v.data {
		if !fn(&ec2.DescribeVpcsOutput{Vpcs: page}, true) {
			return fmt.Errorf("should always return true")
		}
	}
	return nil
}
