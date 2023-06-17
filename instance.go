package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/labstack/echo/v4"
)

var errNotFound = errors.New("Not found")

type InstanceClient interface {
	LaunchInstance(c echo.Context) (*Target, error)
	FindInstance(c echo.Context) (*Target, error)
	CheckInstance(instance any) (bool, error)
}

type AlarmClient interface {
	AutoTerminate(c echo.Context, t *Target) error
}

var (
	_ InstanceClient = &Ec2Client{}
	_ AlarmClient    = &Ec2AlarmClient{}
)

type Ec2Client = Ec2Config

func NewInstanceClientFromConfig(c *Config) InstanceClient {
	return &c.Ec2Config
}

func (ec *Ec2Client) getSvc() (*ec2.EC2, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(ec.Region),
	})
	if err != nil {
		return nil, fmt.Errorf("failed create session: %w", err)
	}

	return ec2.New(sess), nil

}

func (ec *Ec2Client) LaunchInstance(c echo.Context) (*Target, error) {
	svc, err := ec.getSvc()
	if err != nil {
		return nil, err
	}

	c.Logger().Debugf("Start instance with script: %s", ec.StartScript)

	input := &ec2.RunInstancesInput{
		ImageId:      aws.String(ec.ImageId),
		InstanceType: aws.String(ec.InstanceType),
		MinCount:     aws.Int64(1),
		MaxCount:     aws.Int64(1),
		KeyName:      aws.String(ec.KeyName),
		UserData:     aws.String(base64.StdEncoding.EncodeToString([]byte(ec.StartScript))),
		SecurityGroupIds: []*string{
			aws.String(ec.SecurityGroupId),
		},
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("instance"),
				Tags: []*ec2.Tag{
					{
						Key:   aws.String("name"),
						Value: aws.String(ec.Tag),
					},
				},
			},
		},
		BlockDeviceMappings: []*ec2.BlockDeviceMapping{
			{
				DeviceName: aws.String("/dev/sda1"),
				Ebs: &ec2.EbsBlockDevice{
					VolumeSize: aws.Int64(ec.DiskSize),
				},
			},
		},
	}

	result, err := svc.RunInstances(input)
	if err != nil {
		return nil, fmt.Errorf("failed to launch instance: %w", err)
	}

	c.Logger().Infof("Created instance: %v", result)

	if len(result.Instances) == 0 {
		return nil, fmt.Errorf("failed to launch instance: empty instance")
	}

	url, err := ec.getInstanceURL(result.Instances[0])

	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %w", err)
	}

	return &Target{
		URL:      url,
		Instance: result.Instances[0],
	}, nil
}

func (ec *Ec2Client) getInstanceURL(instance *ec2.Instance) (*url.URL, error) {
	var (
		url *url.URL
		err error
	)

	if ec.UsePrivateDns {
		url, err = url.Parse(fmt.Sprintf("http://%s", *instance.PrivateDnsName))
	} else {
		url, err = url.Parse(fmt.Sprintf("http://%s", *instance.PublicIpAddress))
	}

	if err != nil {
		return nil, err
	}

	if ec.Port > 0 {
		url.Host = fmt.Sprintf("%v:%v", url.Host, ec.Port)
	}

	return url, nil
}

func (ec *Ec2Client) FindInstance(c echo.Context) (*Target, error) {
	svc, err := ec.getSvc()
	if err != nil {
		return nil, err
	}

	input := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:name"),
				Values: []*string{
					aws.String(ec.Tag),
				},
			},
		},
	}

	result, err := svc.DescribeInstances(input)
	if err != nil {
		return nil, fmt.Errorf("failed to find instance: %w", err)
	}

	for _, r := range result.Reservations {
		for _, instance := range r.Instances {
			c.Logger().Infof("Found instance: %v", instance)

			if *instance.State.Name == "running" {
				url, err := ec.getInstanceURL(instance)
				if err != nil {
					return nil, err
				}

				return &Target{
					URL:      url,
					Instance: instance,
				}, nil
			}
		}
	}

	return nil, errNotFound
}

func (ec *Ec2Client) CheckInstance(instance any) (bool, error) {
	i, ok := instance.(*ec2.Instance)
	if !ok {
		return false, errors.New("not EC2 instance type")
	}

	svc, err := ec.getSvc()
	if err != nil {
		return false, err
	}

	input := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			i.InstanceId,
		},
	}

	result, err := svc.DescribeInstances(input)
	if err != nil {
		return false, fmt.Errorf("failed to describe instances: %v", err)
	}

	if len(result.Reservations) == 0 || len(result.Reservations[0].Instances) == 0 {
		return false, nil
	}

	i = result.Reservations[0].Instances[0]

	if aws.StringValue(i.State.Name) == ec2.InstanceStateNameRunning {
		return true, nil
	}

	return false, nil
}

type Ec2AlarmClient = Ec2Config

func NewAlarmClientFromConfig(config *Config) AlarmClient {
	return &config.Ec2Config
}

func (ea *Ec2AlarmClient) AutoTerminate(c echo.Context, t *Target) error {
	instanceID := t.Instance.InstanceId
	alarmName := fmt.Sprintf("AutoTermintate-%s", *instanceID)

	c.Logger().Debugf("Setting alarm with region %s", ea.Region)

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(ea.Region),
	})

	if err != nil {
		return err
	}

	svc := cloudwatch.New(sess)

	input := &cloudwatch.PutMetricAlarmInput{
		AlarmName:          aws.String(alarmName),
		ComparisonOperator: aws.String(cloudwatch.ComparisonOperatorLessThanThreshold),
		EvaluationPeriods:  aws.Int64(10),
		MetricName:         aws.String("CPUUtilization"),
		Namespace:          aws.String("AWS/EC2"),
		Period:             aws.Int64(60),
		Statistic:          aws.String(cloudwatch.StatisticMaximum),
		Threshold:          aws.Float64(2.0),
		ActionsEnabled:     aws.Bool(true),
		AlarmDescription:   aws.String("Terminate instance if CPU utilization is below 5% for 10 minutes"),
		Unit:               aws.String(cloudwatch.StandardUnitPercent),
		Dimensions: []*cloudwatch.Dimension{
			{
				Name:  aws.String("InstanceId"),
				Value: instanceID,
			},
		},
		AlarmActions: []*string{
			aws.String(fmt.Sprintf("arn:aws:automate:%s:ec2:terminate", ea.Region)),
		},
	}

	_, err = svc.PutMetricAlarm(input)
	if err != nil {
		return err
	}

	return nil
}
