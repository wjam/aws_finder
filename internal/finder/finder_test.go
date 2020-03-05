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

	"github.com/aws/aws-sdk-go/aws/request"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchPerRegion_Config(t *testing.T) {
	configFile, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	defer os.Remove(configFile.Name())
	err = ioutil.WriteFile(configFile.Name(), []byte(`
[profile default]
foo = bar

[profile dev]
foo = baz

[profile prod]
foo = qux

[profile region-failure]
this = will-fail
`), 0600)
	require.NoError(t, err)
	existingCredentials := os.Getenv("AWS_SHARED_CREDENTIALS_FILE")
	existingConfig := os.Getenv("AWS_CONFIG_FILE")
	defer os.Setenv("AWS_CONFIG_FILE", existingConfig)
	defer os.Setenv("AWS_CONFIG_FILE", existingCredentials)
	err = os.Setenv("AWS_CONFIG_FILE", configFile.Name())
	require.NoError(t, err)

	osUserHomeDir = func() (string, error) {
		return ioutil.TempDir("", "")
	}
	ec2New = func(c client.ConfigProvider) regionLister {
		if sess, ok := c.(*session.Session); ok {
			if aws.StringValue(sess.Config.Endpoint) == "region-failure" {
				return &rFailure{}
			}
		}
		return &r{}
	}
	newSession = func(region string, profile string) *session.Session {
		ret := &session.Session{
			Config: &aws.Config{
				Region:   aws.String(region),
				Endpoint: aws.String(profile),
			},
		}
		return ret
	}

	var lock sync.RWMutex
	var prefixes []string
	SearchPerRegion(context.TODO(), func(ctx context.Context, l *log.Logger, _ *session.Session) {
		lock.Lock()
		defer lock.Unlock()
		prefixes = append(prefixes, strings.TrimSpace(l.Prefix()))
	})

	require.Len(t, prefixes, 9)
	assert.Contains(t, prefixes, "[default] [eu-west-1]")
	assert.Contains(t, prefixes, "[default] [eu-west-2]")
	assert.Contains(t, prefixes, "[default] [us-east-1]")
	assert.Contains(t, prefixes, "[dev] [eu-west-1]")
	assert.Contains(t, prefixes, "[dev] [eu-west-2]")
	assert.Contains(t, prefixes, "[dev] [us-east-1]")
	assert.Contains(t, prefixes, "[prod] [eu-west-1]")
	assert.Contains(t, prefixes, "[prod] [eu-west-2]")
	assert.Contains(t, prefixes, "[prod] [us-east-1]")
}

func TestSearch_Credentials(t *testing.T) {
	credentialsFile, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	defer os.Remove(credentialsFile.Name())
	err = ioutil.WriteFile(credentialsFile.Name(), []byte(`
[default]
aws_access_key_id=123
aws_secret_access_key=321

[dev]
aws_access_key_id=123
aws_secret_access_key=321

[baz]
aws_access_key_id=123
aws_secret_access_key=321
`), 0600)
	require.NoError(t, err)
	existingCredentials := os.Getenv("AWS_SHARED_CREDENTIALS_FILE")
	existingConfig := os.Getenv("AWS_CONFIG_FILE")
	defer os.Setenv("AWS_CONFIG_FILE", existingConfig)
	defer os.Setenv("AWS_CONFIG_FILE", existingCredentials)
	err = os.Setenv("AWS_SHARED_CREDENTIALS_FILE", credentialsFile.Name())
	require.NoError(t, err)

	osUserHomeDir = func() (string, error) {
		return ioutil.TempDir("", "")
	}
	ec2New = func(client.ConfigProvider) regionLister {
		return &r{}
	}
	newSession = func(string, string) *session.Session {
		return nil
	}

	var lock sync.RWMutex
	var prefixes []string
	SearchPerRegion(context.TODO(), func(ctx context.Context, l *log.Logger, _ *session.Session) {
		lock.Lock()
		defer lock.Unlock()
		prefixes = append(prefixes, strings.TrimSpace(l.Prefix()))
	})

	require.Len(t, prefixes, 9)
	assert.Contains(t, prefixes, "[default] [eu-west-1]")
	assert.Contains(t, prefixes, "[default] [eu-west-2]")
	assert.Contains(t, prefixes, "[default] [us-east-1]")
	assert.Contains(t, prefixes, "[dev] [eu-west-1]")
	assert.Contains(t, prefixes, "[dev] [eu-west-2]")
	assert.Contains(t, prefixes, "[dev] [us-east-1]")
	assert.Contains(t, prefixes, "[baz] [eu-west-1]")
	assert.Contains(t, prefixes, "[baz] [eu-west-2]")
	assert.Contains(t, prefixes, "[baz] [us-east-1]")
}

func TestSearchPerRegion_ConfigAndCredentials(t *testing.T) {
	credentialsFile, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	defer os.Remove(credentialsFile.Name())

	configFile, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	defer os.Remove(configFile.Name())

	err = ioutil.WriteFile(credentialsFile.Name(), []byte(`
[default]
aws_access_key_id=123
aws_secret_access_key=321

[dev]
aws_access_key_id=123
aws_secret_access_key=321

[baz]
aws_access_key_id=123
aws_secret_access_key=321
`), 0600)
	require.NoError(t, err)
	err = ioutil.WriteFile(configFile.Name(), []byte(`
[profile dev]
foo = baz

[profile prod]
foo = qux
`), 0600)
	require.NoError(t, err)

	existingCredentials := os.Getenv("AWS_SHARED_CREDENTIALS_FILE")
	defer os.Setenv("AWS_SHARED_CREDENTIALS_FILE", existingCredentials)
	err = os.Setenv("AWS_SHARED_CREDENTIALS_FILE", credentialsFile.Name())
	require.NoError(t, err)
	existingConfig := os.Getenv("AWS_CONFIG_FILE")
	defer os.Setenv("AWS_CONFIG_FILE", existingConfig)
	err = os.Setenv("AWS_CONFIG_FILE", configFile.Name())
	require.NoError(t, err)

	osUserHomeDir = func() (string, error) {
		return ioutil.TempDir("", "")
	}
	ec2New = func(client.ConfigProvider) regionLister {
		return &r{}
	}
	newSession = func(string, string) *session.Session {
		return nil
	}

	var lock sync.RWMutex
	var prefixes []string
	SearchPerRegion(context.TODO(), func(ctx context.Context, l *log.Logger, _ *session.Session) {
		lock.Lock()
		defer lock.Unlock()
		prefixes = append(prefixes, strings.TrimSpace(l.Prefix()))
	})

	require.Len(t, prefixes, 12)
	assert.Contains(t, prefixes, "[default] [eu-west-1]")
	assert.Contains(t, prefixes, "[default] [eu-west-2]")
	assert.Contains(t, prefixes, "[default] [us-east-1]")
	assert.Contains(t, prefixes, "[dev] [eu-west-1]")
	assert.Contains(t, prefixes, "[dev] [eu-west-2]")
	assert.Contains(t, prefixes, "[dev] [us-east-1]")
	assert.Contains(t, prefixes, "[baz] [eu-west-1]")
	assert.Contains(t, prefixes, "[baz] [eu-west-2]")
	assert.Contains(t, prefixes, "[baz] [us-east-1]")
	assert.Contains(t, prefixes, "[prod] [eu-west-1]")
	assert.Contains(t, prefixes, "[prod] [eu-west-2]")
	assert.Contains(t, prefixes, "[prod] [us-east-1]")
}

func TestSearchPerProfile_Config(t *testing.T) {
	configFile, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	defer os.Remove(configFile.Name())
	err = ioutil.WriteFile(configFile.Name(), []byte(`
[profile default]
foo = bar

[profile dev]
foo = baz

[profile prod]
foo = qux

[profile region-failure]
this = will-fail
`), 0600)
	require.NoError(t, err)
	existingCredentials := os.Getenv("AWS_SHARED_CREDENTIALS_FILE")
	existingConfig := os.Getenv("AWS_CONFIG_FILE")
	defer os.Setenv("AWS_CONFIG_FILE", existingConfig)
	defer os.Setenv("AWS_CONFIG_FILE", existingCredentials)
	err = os.Setenv("AWS_CONFIG_FILE", configFile.Name())
	require.NoError(t, err)

	osUserHomeDir = func() (string, error) {
		return ioutil.TempDir("", "")
	}
	newSession = func(region string, profile string) *session.Session {
		ret := &session.Session{
			Config: &aws.Config{
				Region:   aws.String(region),
				Endpoint: aws.String(profile),
			},
		}
		return ret
	}

	var lock sync.RWMutex
	var prefixes []string
	SearchPerProfile(context.TODO(), func(ctx context.Context, l *log.Logger, _ *session.Session) {
		lock.Lock()
		defer lock.Unlock()
		prefixes = append(prefixes, strings.TrimSpace(l.Prefix()))
	})

	require.Len(t, prefixes, 4)
	assert.Contains(t, prefixes, "[default]")
	assert.Contains(t, prefixes, "[dev]")
	assert.Contains(t, prefixes, "[prod]")
	assert.Contains(t, prefixes, "[region-failure]")
}

var _ regionLister = &r{}

type rFailure struct {
}

func (r *rFailure) DescribeRegionsWithContext(ctx aws.Context, input *ec2.DescribeRegionsInput, opts ...request.Option) (*ec2.DescribeRegionsOutput, error) {
	return nil, fmt.Errorf("something went wrong")
}

type r struct {
}

func (r *r) DescribeRegionsWithContext(ctx aws.Context, input *ec2.DescribeRegionsInput, opts ...request.Option) (*ec2.DescribeRegionsOutput, error) {
	return &ec2.DescribeRegionsOutput{
		Regions: []*ec2.Region{
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
