package finder

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wjam/aws_finder/internal/log"
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
	t.Setenv("AWS_CONFIG_FILE", configFile)

	osUserHomeDir = func() (string, error) {
		return t.TempDir(), nil
	}
	ec2New = func(c aws.Config) regionLister {
		if c.ConfigSources[0] == "region-failure" {
			return &rFailure{}
		}
		return &r{}
	}
	newSession = func(_ context.Context, region, profile string) (aws.Config, error) {
		ret := aws.Config{
			Region:        region,
			ConfigSources: []interface{}{profile},
		}
		return ret, nil
	}

	var lock sync.RWMutex
	capture := captureHandler{}

	ctx := log.ContextWithLogger(t.Context(), slog.New(log.WithAttrsFromContextHandler{
		Parent: &capture,
	}))

	assert.Error(t, SearchPerRegion(ctx, func(ctx context.Context, _ aws.Config) error {
		lock.Lock()
		defer lock.Unlock()

		log.Logger(ctx).InfoContext(ctx, "log")
		return nil
	}))

	assert.ElementsMatch(t, *capture.r, []string{
		"profile=default region=eu-west-1",
		"profile=default region=eu-west-2",
		"profile=default region=us-east-1",
		"profile=dev region=eu-west-1",
		"profile=dev region=eu-west-2",
		"profile=dev region=us-east-1",
		"profile=prod region=eu-west-1",
		"profile=prod region=eu-west-2",
		"profile=prod region=us-east-1",
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
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", credentialsFile)

	osUserHomeDir = func() (string, error) {
		return t.TempDir(), nil
	}
	ec2New = func(_ aws.Config) regionLister {
		return &r{}
	}
	newSession = func(_ context.Context, _, _ string) (aws.Config, error) {
		return aws.Config{}, nil
	}

	var lock sync.RWMutex
	capture := captureHandler{}

	ctx := log.ContextWithLogger(t.Context(), slog.New(log.WithAttrsFromContextHandler{
		Parent: &capture,
	}))

	err := SearchPerRegion(ctx, func(ctx context.Context, _ aws.Config) error {
		lock.Lock()
		defer lock.Unlock()

		log.Logger(ctx).InfoContext(ctx, "log")
		return nil
	})
	require.NoError(t, err)

	assert.ElementsMatch(t, *capture.r, []string{
		"profile=default region=eu-west-1",
		"profile=default region=eu-west-2",
		"profile=default region=us-east-1",
		"profile=dev region=eu-west-1",
		"profile=dev region=eu-west-2",
		"profile=dev region=us-east-1",
		"profile=baz region=eu-west-1",
		"profile=baz region=eu-west-2",
		"profile=baz region=us-east-1",
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
	t.Setenv("AWS_CONFIG_FILE", configFile)
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", credentialsFile)

	osUserHomeDir = func() (string, error) {
		return t.TempDir(), nil
	}
	ec2New = func(_ aws.Config) regionLister {
		return &r{}
	}
	newSession = func(_ context.Context, _, _ string) (aws.Config, error) {
		return aws.Config{}, nil
	}

	var lock sync.RWMutex
	capture := captureHandler{}

	ctx := log.ContextWithLogger(t.Context(), slog.New(log.WithAttrsFromContextHandler{
		Parent: &capture,
	}))

	err := SearchPerRegion(ctx, func(ctx context.Context, _ aws.Config) error {
		lock.Lock()
		defer lock.Unlock()

		log.Logger(ctx).InfoContext(ctx, "log")
		return nil
	})
	require.NoError(t, err)

	assert.ElementsMatch(t, *capture.r, []string{
		"profile=default region=eu-west-1",
		"profile=default region=eu-west-2",
		"profile=default region=us-east-1",
		"profile=dev region=eu-west-1",
		"profile=dev region=eu-west-2",
		"profile=dev region=us-east-1",
		"profile=baz region=eu-west-1",
		"profile=baz region=eu-west-2",
		"profile=baz region=us-east-1",
		"profile=prod region=eu-west-1",
		"profile=prod region=eu-west-2",
		"profile=prod region=us-east-1",
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
	t.Setenv("AWS_CONFIG_FILE", configFile)

	osUserHomeDir = func() (string, error) {
		return t.TempDir(), nil
	}
	newSession = func(_ context.Context, region, profile string) (aws.Config, error) {
		ret := aws.Config{
			Region:        region,
			ConfigSources: []interface{}{profile},
		}
		return ret, nil
	}

	var lock sync.RWMutex
	capture := captureHandler{}

	ctx := log.ContextWithLogger(t.Context(), slog.New(log.WithAttrsFromContextHandler{
		Parent: &capture,
	}))

	err := SearchPerProfile(ctx, func(ctx context.Context, _ aws.Config) error {
		lock.Lock()
		defer lock.Unlock()

		log.Logger(ctx).InfoContext(ctx, "log")
		return nil
	})
	require.NoError(t, err)

	require.Len(t, *capture.r, 4)
	assert.Contains(t, *capture.r, "profile=default")
	assert.Contains(t, *capture.r, "profile=dev")
	assert.Contains(t, *capture.r, "profile=prod")
	assert.Contains(t, *capture.r, "profile=region-failure")
}

var _ regionLister = &r{}
var _ regionLister = &rFailure{}

type rFailure struct {
}

func (r *rFailure) DescribeRegions(
	_ context.Context, _ *ec2.DescribeRegionsInput, _ ...func(*ec2.Options),
) (*ec2.DescribeRegionsOutput, error) {
	return nil, errors.New("something went wrong")
}

type r struct {
}

func (r *r) DescribeRegions(
	_ context.Context, _ *ec2.DescribeRegionsInput, _ ...func(*ec2.Options),
) (*ec2.DescribeRegionsOutput, error) {
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

func tempFile(t *testing.T, content string) string {
	dir := t.TempDir()

	credentialFile := filepath.Join(dir, "credentials")

	require.NoError(t, os.WriteFile(credentialFile, []byte(content), 0600))

	return credentialFile
}

var _ io.Writer = noOpWriter{}

type noOpWriter struct {
}

func (n noOpWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

type captureHandler struct {
	attrs []slog.Attr
	r     *[]string
}

func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	var vals []string

	for _, attr := range h.attrs {
		vals = append(vals, attr.String())
	}

	r.Attrs(func(attr slog.Attr) bool {
		vals = append(vals, attr.String())
		return true
	})

	allVals := []string{strings.Join(vals, " ")}
	if h.r != nil {
		allVals = append(*h.r, allVals...)
	}
	h.r = &allVals
	return nil
}

func (*captureHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *captureHandler) WithAttrs(as []slog.Attr) slog.Handler {
	var c2 captureHandler
	c2.r = h.r

	//nolint:gocritic // we're copying the attributes to the child handler
	c2.attrs = append(h.attrs, as...)
	return &c2
}

func (*captureHandler) WithGroup(string) slog.Handler {
	panic("not implemented")
}
