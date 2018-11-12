package cvm

import (
	"time"
)

const (
	FilterNameZone               = "zone"
	FilterNameProjectId          = "project-id"
	FilterNameHostId             = "host-id"
	FilterNameInstanceId         = "instance-id"
	FilterNameInstanceName       = "instance-name"
	FilterNameInstanceChargeType = "instance-charge-type"
	FilterNamePrivateIpAddress   = "private-ip-address"
	FilterNamePublicIpAddress    = "public-ip-address"

	DefaultVersion = "2017-03-12"
)

type DescribeInstancesArgs struct {
	Version     string    `qcloud_arg:"Version,required"`
	InstanceIds *[]string `qcloud_arg:"InstanceIds"`
	Filters     *[]Filter `qcloud_arg:"Filters"`
	Offset      *int      `qcloud_arg:"Offset"`
	Limit       *int      `qcloud_arg:"Limit"`
}

type RunInstancesArgs struct{
	Placements   Placement  `qcloud_arg:"Placement"`
	ImageId		string	`qcloud_arg:"ImageId"`
	InternetAccessibles		InternetAccessible		`qcloud_arg:"InternetAccessible"`
	//Cpu			int 	`qcloud_arg:"cpu"`
	//Mem			int		`qcloud_arg:"mem"`
	//ZoneId		int		`qcloud_arg:"zoneId"`
	//StorageSize	int		`qcloud_arg:"storageSize"`
}

type Placement struct {
	Zone			string 		`qcloud_arg:"Zone"`
	//ProjectId		int			`qcloud_arg:"ProjectId"`
	//HostIds			[]string	`qcloud_arg:"HostIds"`
}

type InternetAccessible struct{
	InternetChargeType			string		`qcloud_arg:"InternetChargeType"`
	InternetMaxBandwidthOut		int			`qcloud_arg:"InternetMaxBandwidthOut"`
	PublicIpAssigned			bool		`qcloud_arg:"PublicIpAssigned"`
}

type Filter struct {
	Name   string        `qcloud_arg:"Name"`
	Values []interface{} `qcloud_arg:"Values"`
}

func NewFilter(name string, values ...interface{}) Filter {
	return Filter{Name: name, Values: values}
}

type CvmResponse struct {
	Response interface{} `json:"Response"`
}

type DescribeInstancesResponse struct {
	TotalCount  int            `json:"TotalCount"`
	InstanceSet []InstanceInfo `json:"InstanceSet"`
	RequestID   string         `json:"RequestId"`
}

type RunInstancesResponse struct{
	RequestId		string			`json:"RequestId"`
	InstanceIdSet 	[]string		`json:"InstanceIdSet"`
}

//type Placement struct {
//	Zone      string      `json:"Zone"`
//	HostID    interface{} `json:"HostId"`
//	ProjectID int         `json:"ProjectId"`
//}

type Disk struct {
	DiskType string `json:"DiskType"`
	DiskID   string `json:"DiskId"`
	DiskSize int    `json:"DiskSize"`
}

//type InternetAccessible struct {
//	InternetMaxBandwidthOut int    `json:"InternetMaxBandwidthOut"`
//	InternetChargeType      string `json:"InternetChargeType"`
//}

type VirtualPrivateCloud struct {
	VpcID        string `json:"VpcId"`
	SubnetID     string `json:"SubnetId"`
	AsVpcGateway bool   `json:"AsVpcGateway"`
}

type InstanceInfo struct {
	InstanceID         string   `json:"InstanceId"`
	InstanceType       string   `json:"InstanceType"`
	CPU                int      `json:"CPU"`
	Memory             int      `json:"Memory"`
	InstanceName       string   `json:"InstanceName"`
	InstanceChargeType string   `json:"InstanceChargeType"`
	PrivateIPAddresses []string `json:"PrivateIpAddresses"`
	PublicIPAddresses  []string `json:"PublicIpAddresses"`
	ImageID            string   `json:"ImageId"`
	OsName             string   `json:"OsName"`
	RenewFlag          string   `json:"RenewFlag"`

	Placement           Placement           `json:"Placement"`
	SystemDisk          Disk                `json:"SystemDisk"`
	DataDisks           []Disk              `json:"DataDisks"`
	InternetAccessible  InternetAccessible  `json:"InternetAccessible"`
	VirtualPrivateCloud VirtualPrivateCloud `json:"VirtualPrivateCloud"`

	CreatedTime time.Time `json:"CreatedTime"`
	ExpiredTime time.Time `json:"ExpiredTime"`
}

func (client *Client) DescribeInstances(args *DescribeInstancesArgs) (*DescribeInstancesResponse, error) {
	realRsp := &DescribeInstancesResponse{}
	cvmResponse := &CvmResponse{
		Response: realRsp,
	}
	err := client.Invoke("DescribeInstances", args, cvmResponse)
	if err != nil {
		return &DescribeInstancesResponse{}, err
	}
	return realRsp, nil
}


func (client *Client) RunInstances(args *RunInstancesArgs) (*RunInstancesResponse, error){
	realRsp := &RunInstancesResponse{}
	err := client.Invoke("RunInstances", args, realRsp)
	if err != nil {
		return &RunInstancesResponse{}, err
	}
	return realRsp, nil
}