package finder

import (
	"io/ioutil"
	"log"
	"os"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearch_Config(t *testing.T) {
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
`), 0600)
	require.NoError(t, err)
	existingCredentials := os.Getenv("AWS_CREDENTIAL_FILE")
	existingConfig := os.Getenv("AWS_CONFIG_FILE")
	defer os.Setenv("AWS_CONFIG_FILE", existingConfig)
	defer os.Setenv("AWS_CONFIG_FILE", existingCredentials)
	err = os.Setenv("AWS_CONFIG_FILE", configFile.Name())
	require.NoError(t, err)

	osUserHomeDir = func() (string, error) {
		return ioutil.TempDir("", "")
	}
	ec2New = func(client.ConfigProvider) regionLister {
		return &r{}
	}
	newSession = func(*aws.Config) *session.Session {
		return nil
	}

	var lock sync.RWMutex
	var prefixes []string
	Search(func(l *log.Logger, _ *session.Session) {
		lock.Lock()
		defer lock.Unlock()
		prefixes = append(prefixes, l.Prefix())
	})

	require.Len(t, prefixes, 9)
	assert.Contains(t, prefixes, "[default][eu-west-1]")
	assert.Contains(t, prefixes, "[default][eu-west-2]")
	assert.Contains(t, prefixes, "[default][us-east-1]")
	assert.Contains(t, prefixes, "[dev][eu-west-1]")
	assert.Contains(t, prefixes, "[dev][eu-west-2]")
	assert.Contains(t, prefixes, "[dev][us-east-1]")
	assert.Contains(t, prefixes, "[prod][eu-west-1]")
	assert.Contains(t, prefixes, "[prod][eu-west-2]")
	assert.Contains(t, prefixes, "[prod][us-east-1]")
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
	existingCredentials := os.Getenv("AWS_CREDENTIAL_FILE")
	existingConfig := os.Getenv("AWS_CONFIG_FILE")
	defer os.Setenv("AWS_CONFIG_FILE", existingConfig)
	defer os.Setenv("AWS_CONFIG_FILE", existingCredentials)
	err = os.Setenv("AWS_CREDENTIAL_FILE", credentialsFile.Name())
	require.NoError(t, err)

	osUserHomeDir = func() (string, error) {
		return ioutil.TempDir("", "")
	}
	ec2New = func(client.ConfigProvider) regionLister {
		return &r{}
	}
	newSession = func(*aws.Config) *session.Session {
		return nil
	}

	var lock sync.RWMutex
	var prefixes []string
	Search(func(l *log.Logger, _ *session.Session) {
		lock.Lock()
		defer lock.Unlock()
		prefixes = append(prefixes, l.Prefix())
	})

	require.Len(t, prefixes, 9)
	assert.Contains(t, prefixes, "[default][eu-west-1]")
	assert.Contains(t, prefixes, "[default][eu-west-2]")
	assert.Contains(t, prefixes, "[default][us-east-1]")
	assert.Contains(t, prefixes, "[dev][eu-west-1]")
	assert.Contains(t, prefixes, "[dev][eu-west-2]")
	assert.Contains(t, prefixes, "[dev][us-east-1]")
	assert.Contains(t, prefixes, "[baz][eu-west-1]")
	assert.Contains(t, prefixes, "[baz][eu-west-2]")
	assert.Contains(t, prefixes, "[baz][us-east-1]")
}

func TestSearch_ConfigAndCredentials(t *testing.T) {
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

	existingCredentials := os.Getenv("AWS_CREDENTIAL_FILE")
	defer os.Setenv("AWS_CREDENTIAL_FILE", existingCredentials)
	err = os.Setenv("AWS_CREDENTIAL_FILE", credentialsFile.Name())
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
	newSession = func(*aws.Config) *session.Session {
		return nil
	}

	var lock sync.RWMutex
	var prefixes []string
	Search(func(l *log.Logger, _ *session.Session) {
		lock.Lock()
		defer lock.Unlock()
		prefixes = append(prefixes, l.Prefix())
	})

	require.Len(t, prefixes, 12)
	assert.Contains(t, prefixes, "[default][eu-west-1]")
	assert.Contains(t, prefixes, "[default][eu-west-2]")
	assert.Contains(t, prefixes, "[default][us-east-1]")
	assert.Contains(t, prefixes, "[dev][eu-west-1]")
	assert.Contains(t, prefixes, "[dev][eu-west-2]")
	assert.Contains(t, prefixes, "[dev][us-east-1]")
	assert.Contains(t, prefixes, "[baz][eu-west-1]")
	assert.Contains(t, prefixes, "[baz][eu-west-2]")
	assert.Contains(t, prefixes, "[baz][us-east-1]")
	assert.Contains(t, prefixes, "[prod][eu-west-1]")
	assert.Contains(t, prefixes, "[prod][eu-west-2]")
	assert.Contains(t, prefixes, "[prod][us-east-1]")
}

var _ regionLister = &r{}

type r struct {
}

func (r *r) DescribeRegions(*ec2.DescribeRegionsInput) (*ec2.DescribeRegionsOutput, error) {
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
