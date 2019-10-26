package finder

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	mapset "github.com/deckarep/golang-set"

	"github.com/aws/aws-sdk-go/aws/client"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"gopkg.in/ini.v1"
)

// Search will call `f` for every region in every profile defined in ~/.aws/config
func Search(f func(*log.Logger, *session.Session)) {
	profiles, err := profiles()
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup

	for _, profile := range profiles.ToSlice() {
		l := log.New(os.Stdout, fmt.Sprintf("[%s]", profile), 0)

		wg.Add(1)
		go func() {
			defer wg.Done()
			searchAccount(profile.(string), l, f)
		}()

		wg.Wait()
	}
}

func searchAccount(profile string, parentLogger *log.Logger, f func(*log.Logger, *session.Session)) {
	regions, err := regions(profile)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	for _, region := range regions {
		sess := newSession(aws.NewConfig().
			WithRegion(region).
			WithCredentials(credentials.NewSharedCredentials("", profile)))

		l := log.New(parentLogger.Writer(), fmt.Sprintf("%s[%s]", parentLogger.Prefix(), region), 0)

		wg.Add(1)
		go func() {
			defer wg.Done()
			f(l, sess)
		}()

		wg.Wait()
	}
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
	config, err := ini.Load(file)
	if err != nil {
		return nil, err
	}

	profiles := mapset.NewSet()
	for _, section := range config.Sections() {
		if !strings.HasPrefix(section.Name(), "profile ") {
			continue
		}
		profiles.Add(strings.TrimPrefix(section.Name(), "profile "))
	}

	return profiles, nil
}

func profilesFromCredentialsFile() (mapset.Set, error) {
	file, err := credentialsFile()
	config, err := ini.Load(file)
	if err != nil {
		return nil, err
	}

	profiles := mapset.NewSet()
	for _, section := range config.Sections() {
		if section.Name() == ini.DefaultSection {
			continue
		}
		profiles.Add(section.Name())
	}

	return profiles, nil
}

func regions(profile string) ([]string, error) {
	sess := newSession(aws.NewConfig().WithCredentials(credentials.NewSharedCredentials("", profile)))

	output, err := ec2New(sess).DescribeRegions(&ec2.DescribeRegionsInput{})
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
	if file, ok := os.LookupEnv("AWS_CREDENTIAL_FILE"); ok {
		return file, nil
	}

	home, err := osUserHomeDir()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/.aws/credentials", home), nil
}

// Things to be neutered when running in tests
var ec2New = func(p client.ConfigProvider) regionLister {
	return ec2.New(p)
}
var newSession = func(c *aws.Config) *session.Session {
	return session.Must(session.NewSession(c))
}
var osUserHomeDir = os.UserHomeDir

type regionLister interface {
	DescribeRegions(*ec2.DescribeRegionsInput) (*ec2.DescribeRegionsOutput, error)
}
