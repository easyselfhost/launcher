package main

import (
	"net/url"

	"github.com/aws/aws-sdk-go/service/ec2"
)

type Target struct {
	URL      *url.URL
	Instance *ec2.Instance
}
