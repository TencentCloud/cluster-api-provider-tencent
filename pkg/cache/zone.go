package cache

import (
	"github.com/pkg/errors"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

var (
	ZoneNameToID = map[string]string{}
)

func InitZoneCache(client *cvm.Client) error {
	if len(ZoneNameToID) == 0 {
		req := cvm.NewDescribeZonesRequest()

		res, err := client.DescribeZones(req)
		if err != nil {
			return errors.Wrap(err, "unable to describe zone")
		}

		for _, zoneInfo := range res.Response.ZoneSet {
			ZoneNameToID[*zoneInfo.Zone] = *zoneInfo.ZoneId
		}

		return nil
	}
	return nil
}
