package util

import (
	"os"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
)

type Credentials struct {
	Region      string
	Credentials common.CredentialIface
	Profile     *profile.ClientProfile
}

func NewCredentialsFromENV() *Credentials {
	return &Credentials{
		Region: regions.Mumbai,
		Credentials: common.NewCredential(
			os.Getenv("TENCENTCLOUD_SECRET_ID"),
			os.Getenv("TENCENTCLOUD_SECRET_KEY"),
		),
		Profile: profile.NewClientProfile(),
	}
}

func (c *Credentials) WithRegion(region string) *Credentials {
	c.Region = region
	return c
}

func (c *Credentials) AsArgs() (common.CredentialIface, string, *profile.ClientProfile) {
	return c.Credentials, c.Region, c.Profile
}
