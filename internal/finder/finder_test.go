package finder

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchPerRegion_Config(t *testing.T) {
	configFile := tempFile(t, `
[profile default]
foo = bar

[profile dev]
foo = baz

[profile prod]
foo = qux

[profile region-failure]
this = will-fail
`)
	setEnv(t, "AWS_CONFIG_FILE", configFile.Name())

	osUserHomeDir = func() (string, error) {
		return ioutil.TempDir("", "")
	}
	ec2New = func(c aws.Config) regionLister {
		if c.ConfigSources[0] == "region-failure" {
			return &rFailure{}
		}
		return &r{}
	}
	newSession = func(ctx context.Context, region, profile string) (aws.Config, error) {
		ret := aws.Config{
			Region:        region,
			ConfigSources: []interface{}{profile},
		}
		return ret, nil
	}

	var lock sync.RWMutex
	var prefixes []string
	assert.Error(t, SearchPerRegion(context.Background(), func(ctx context.Context, l *log.Logger, _ aws.Config) error {
		lock.Lock()
		defer lock.Unlock()
		prefixes = append(prefixes, strings.TrimSpace(l.Prefix()))
		return nil
	}))

	assert.ElementsMatch(t, prefixes, []string{
		"[default] [eu-west-1]",
		"[default] [eu-west-2]",
		"[default] [us-east-1]",
		"[dev] [eu-west-1]",
		"[dev] [eu-west-2]",
		"[dev] [us-east-1]",
		"[prod] [eu-west-1]",
		"[prod] [eu-west-2]",
		"[prod] [us-east-1]",
	})

}

func TestSearch_Credentials(t *testing.T) {
	credentialsFile := tempFile(t, `
[default]
aws_access_key_id=123
aws_secret_access_key=321

[dev]
aws_access_key_id=123
aws_secret_access_key=321

[baz]
aws_access_key_id=123
aws_secret_access_key=321
`)
	setEnv(t, "AWS_SHARED_CREDENTIALS_FILE", credentialsFile.Name())

	osUserHomeDir = func() (string, error) {
		return ioutil.TempDir("", "")
	}
	ec2New = func(c aws.Config) regionLister {
		return &r{}
	}
	newSession = func(ctx context.Context, region, profile string) (aws.Config, error) {
		return aws.Config{}, nil
	}

	var lock sync.RWMutex
	var prefixes []string
	require.NoError(t, SearchPerRegion(context.Background(), func(ctx context.Context, l *log.Logger, _ aws.Config) error {
		lock.Lock()
		defer lock.Unlock()
		prefixes = append(prefixes, strings.TrimSpace(l.Prefix()))
		return nil
	}))

	assert.ElementsMatch(t, prefixes, []string{
		"[default] [eu-west-1]",
		"[default] [eu-west-2]",
		"[default] [us-east-1]",
		"[dev] [eu-west-1]",
		"[dev] [eu-west-2]",
		"[dev] [us-east-1]",
		"[baz] [eu-west-1]",
		"[baz] [eu-west-2]",
		"[baz] [us-east-1]",
	})
}

func TestSearchPerRegion_ConfigAndCredentials(t *testing.T) {
	credentialsFile := tempFile(t, `
[default]
aws_access_key_id=123
aws_secret_access_key=321

[dev]
aws_access_key_id=123
aws_secret_access_key=321

[baz]
aws_access_key_id=123
aws_secret_access_key=321
`)
	configFile := tempFile(t, `
[profile dev]
foo = baz

[profile prod]
foo = qux
`)
	setEnv(t, "AWS_CONFIG_FILE", configFile.Name())
	setEnv(t, "AWS_SHARED_CREDENTIALS_FILE", credentialsFile.Name())

	osUserHomeDir = func() (string, error) {
		return ioutil.TempDir("", "")
	}
	ec2New = func(c aws.Config) regionLister {
		return &r{}
	}
	newSession = func(ctx context.Context, region, profile string) (aws.Config, error) {
		return aws.Config{}, nil
	}

	var lock sync.RWMutex
	var prefixes []string
	require.NoError(t, SearchPerRegion(context.Background(), func(ctx context.Context, l *log.Logger, _ aws.Config) error {
		lock.Lock()
		defer lock.Unlock()
		prefixes = append(prefixes, strings.TrimSpace(l.Prefix()))
		return nil
	}))

	assert.ElementsMatch(t, prefixes, []string{
		"[default] [eu-west-1]",
		"[default] [eu-west-2]",
		"[default] [us-east-1]",
		"[dev] [eu-west-1]",
		"[dev] [eu-west-2]",
		"[dev] [us-east-1]",
		"[baz] [eu-west-1]",
		"[baz] [eu-west-2]",
		"[baz] [us-east-1]",
		"[prod] [eu-west-1]",
		"[prod] [eu-west-2]",
		"[prod] [us-east-1]",
	})
}

func TestSearchPerProfile_Config(t *testing.T) {
	configFile := tempFile(t, `
[profile default]
foo = bar

[profile dev]
foo = baz

[profile prod]
foo = qux

[profile region-failure]
this = will-fail
`)
	setEnv(t, "AWS_CONFIG_FILE", configFile.Name())

	osUserHomeDir = func() (string, error) {
		return ioutil.TempDir("", "")
	}
	newSession = func(ctx context.Context, region, profile string) (aws.Config, error) {
		ret := aws.Config{
			Region:        region,
			ConfigSources: []interface{}{profile},
		}
		return ret, nil
	}

	var lock sync.RWMutex
	var prefixes []string
	require.NoError(t, SearchPerProfile(context.Background(), func(ctx context.Context, l *log.Logger, _ aws.Config) error {
		lock.Lock()
		defer lock.Unlock()
		prefixes = append(prefixes, strings.TrimSpace(l.Prefix()))
		return nil
	}))

	require.Len(t, prefixes, 4)
	assert.Contains(t, prefixes, "[default]")
	assert.Contains(t, prefixes, "[dev]")
	assert.Contains(t, prefixes, "[prod]")
	assert.Contains(t, prefixes, "[region-failure]")
}

var _ regionLister = &r{}
var _ regionLister = &rFailure{}

type rFailure struct {
}

func (r *rFailure) DescribeRegions(_ context.Context, _ *ec2.DescribeRegionsInput, _ ...func(*ec2.Options)) (*ec2.DescribeRegionsOutput, error) {
	return nil, fmt.Errorf("something went wrong")
}

type r struct {
}

func (r *r) DescribeRegions(_ context.Context, _ *ec2.DescribeRegionsInput, _ ...func(*ec2.Options)) (*ec2.DescribeRegionsOutput, error) {
	return &ec2.DescribeRegionsOutput{
		Regions: []types.Region{
			{
				RegionName: aws.String("eu-west-1"),
			},
			{
				RegionName: aws.String("eu-west-2"),
			},
			{
				RegionName: aws.String("us-east-1"),
			},
		},
	}, nil
}

func tempFile(t *testing.T, content string) *os.File {
	credentialsFile, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	t.Cleanup(func() {
		err := os.Remove(credentialsFile.Name())
		assert.NoError(t, err)
	})

	err = ioutil.WriteFile(credentialsFile.Name(), []byte(content), 0600)
	require.NoError(t, err)

	return credentialsFile
}

func setEnv(t *testing.T, key, value string) {
	existing := os.Getenv(key)
	t.Cleanup(func() {
		err := os.Setenv(key, existing)
		assert.NoError(t, err)
	})
	err := os.Setenv(key, value)
	require.NoError(t, err)
}
