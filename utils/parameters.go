package utils

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

const prefix = "/moviesapp/appconfig"

var client *ssm.Client

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	Check(err)
	client = ssm.NewFromConfig(cfg)
}

type AppConfigParameters struct {
	AppId           string
	ConfigProfileId string
	EnvId           string
}

func GetParameters() AppConfigParameters {
	a := make([]string, 3)
	a[0] = fmt.Sprintf("%s/application-id", prefix)
	a[1] = fmt.Sprintf("%s/configuration-profile-id", prefix)
	a[2] = fmt.Sprintf("%s/environment-id", prefix)
	input := &ssm.GetParametersInput{
		Names: a,
	}
	o, err := client.GetParameters(context.TODO(), input)
	Check(err)

	return AppConfigParameters{
		AppId:           *o.Parameters[0].Value,
		ConfigProfileId: *o.Parameters[1].Value,
		EnvId:           *o.Parameters[2].Value,
	}
}
