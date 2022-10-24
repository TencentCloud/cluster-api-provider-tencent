package util

import (
	"os"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
)

const (
	TencentCloudSecretID  = "TENCENTCLOUD_SECRET_ID"
	TencentCloudSecretKey = "TENCENTCLOUD_SECRET_KEY"
	LanguageCodeEnUs      = "en-US"
)

type Credentials struct {
	Region      string
	Credentials common.CredentialIface
	Profile     *profile.ClientProfile
}

func NewCredentialsFromENV() *Credentials {
	credentials := &Credentials{
		Region: regions.Mumbai,
		Credentials: common.NewCredential(
			os.Getenv(TencentCloudSecretID),
			os.Getenv(TencentCloudSecretKey),
		),
		Profile: profile.NewClientProfile(),
	}
	credentials.Profile.Language = LanguageCodeEnUs
	return credentials
}

func (c *Credentials) WithRegion(region string) *Credentials {
	c.Region = region
	return c
}

func (c *Credentials) AsArgs() (common.CredentialIface, string, *profile.ClientProfile) {
	return c.Credentials, c.Region, c.Profile
}
