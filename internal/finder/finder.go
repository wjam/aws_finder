package finder

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"strings"
	"sync"

	mapset "github.com/deckarep/golang-set"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"gopkg.in/ini.v1"
)

// SearchPerRegion will call `f` for every region in every profile defined in ~/.aws/config or ~/.aws/credentials
func SearchPerRegion(ctx context.Context, f func(context.Context, *log.Logger, aws.Config)) {
	perProfile(ctx, func(ctx context.Context, profile string, l *log.Logger) {
		perRegion(ctx, profile, l, f)
	})
}

// SearchPerProfile will call `f` for every profile defined in ~/.aws/config or ~/.aws/credentials
func SearchPerProfile(ctx context.Context, f func(context.Context, *log.Logger, aws.Config)) {
	perProfile(ctx, func(ctx context.Context, profile string, l *log.Logger) {
		sess, err := newSession(ctx, "eu-west-1", profile)
		if err != nil {
			panic(err)
		}
		f(ctx, l, sess)
	})
}

func perProfile(ctx context.Context, f func(context.Context, string, *log.Logger)) {
	profiles, err := profiles()
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup

	for _, profile := range profiles.ToSlice() {
		// Shadow the for variable so that it's no longer a pointer, which will change before the go function is run
		profile := profile.(string)
		l := log.New(os.Stdout, fmt.Sprintf("[%s] ", profile), 0)

		wg.Add(1)
		go func() {
			pprof.Do(ctx, pprof.Labels("profile", profile), func(ctx context.Context) {
				defer wg.Done()
				f(ctx, profile, l)
			})
		}()

	}
	wg.Wait()
}

func perRegion(ctx context.Context, profile string, parentLogger *log.Logger, f func(context.Context, *log.Logger, aws.Config)) {
	regions, err := regions(ctx, profile)
	if err != nil {
		parentLogger.Printf("Failed to lookup regions: %s", err)
		return
	}

	var wg sync.WaitGroup
	for _, region := range regions {
		sess, err := newSession(ctx, region, profile)
		if err != nil {
			parentLogger.Printf("Failed to create session for %s: %s", region, err)
			continue
		}

		l := log.New(parentLogger.Writer(), fmt.Sprintf("%s[%s] ", parentLogger.Prefix(), region), 0)
		labelSet := pprof.Labels("region", region)

		wg.Add(1)
		go func() {
			pprof.Do(ctx, labelSet, func(ctx context.Context) {
				defer wg.Done()
				f(ctx, l, sess)
			})
		}()

	}
	wg.Wait()
}

func profiles() (mapset.Set, error) {
	configProfiles, err := profilesFromConfigFile()
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	credentialsProfiles, err := profilesFromCredentialsFile()
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	profiles := mapset.NewSet()
	if configProfiles != nil {
		profiles = profiles.Union(configProfiles)
	}
	if credentialsProfiles != nil {
		profiles = profiles.Union(credentialsProfiles)
	}
	return profiles, nil
}

func profilesFromConfigFile() (mapset.Set, error) {
	file, err := configFile()
	parsed, err := ini.Load(file)
	if err != nil {
		return nil, err
	}

	profiles := mapset.NewSet()
	for _, section := range parsed.Sections() {
		if !strings.HasPrefix(section.Name(), "profile ") {
			continue
		}
		profiles.Add(strings.TrimPrefix(section.Name(), "profile "))
	}

	return profiles, nil
}

func profilesFromCredentialsFile() (mapset.Set, error) {
	file, err := credentialsFile()
	parsed, err := ini.Load(file)
	if err != nil {
		return nil, err
	}

	profiles := mapset.NewSet()
	for _, section := range parsed.Sections() {
		if section.Name() == ini.DefaultSection {
			continue
		}
		profiles.Add(section.Name())
	}

	return profiles, nil
}

func regions(ctx context.Context, profile string) ([]string, error) {
	sess, err := newSession(ctx, "eu-west-1", profile)
	if err != nil {
		return nil, err
	}

	output, err := ec2New(sess).DescribeRegions(ctx, &ec2.DescribeRegionsInput{})
	if err != nil {
		return nil, err
	}

	var regions []string
	for _, region := range output.Regions {
		regions = append(regions, *region.RegionName)
	}

	return regions, nil
}

func configFile() (string, error) {
	if file, ok := os.LookupEnv("AWS_CONFIG_FILE"); ok {
		return file, nil
	}

	home, err := osUserHomeDir()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/.aws/config", home), nil
}

func credentialsFile() (string, error) {
	if file, ok := os.LookupEnv("AWS_SHARED_CREDENTIALS_FILE"); ok {
		return file, nil
	}

	home, err := osUserHomeDir()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/.aws/credentials", home), nil
}

// Things to be neutered when running in tests
var ec2New = func(cfg aws.Config) regionLister {
	return ec2.NewFromConfig(cfg)
}
var newSession = func(ctx context.Context, region, profile string) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region), config.WithSharedConfigProfile(profile))
	if err != nil {
		return aws.Config{}, err
	}
	return cfg, nil
}
var osUserHomeDir = os.UserHomeDir

type regionLister interface {
	DescribeRegions(ctx context.Context, params *ec2.DescribeRegionsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeRegionsOutput, error)
}
