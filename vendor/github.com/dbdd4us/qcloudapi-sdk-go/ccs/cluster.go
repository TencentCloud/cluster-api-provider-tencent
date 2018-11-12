package ccs

type DescribeClustersArgs struct {
	ClusterIds  []string `qcloud_arg:"ClusterIds"`
	ClusterName string	`qcloud_arg:"ClusterName"`
	Offset      *int      `qcloud_arg:"Offset"`
	Limit       *int      `qcloud_arg:"Limit"`
}

type DescribeClustersResponse struct{
	TotalCount		string			`json:"totalCount"`
	Clusters 	[]string			`json:"Clusters"`
	Message			string			`json:"Message"`
}


type CreateClusterArgs struct {
	ClusterName string	`qcloud_arg:"clusterName"`
	ImageId		string	`qcloud_arg:"imageId"`
	Bandwidth	int		`qcloud_arg:"bandwidth"`
	Cpu			int		`qcloud_arg:"cpu"`
	Mem			int		`qcloud_arg:"mem"`
	StorageSize	int		`qcloud_arg:"storageSize"`
	GoodsNum	int 	`qcloud_arg:"goodsNum"`
	ZoneId		string 	`qcloud_arg:"zoneId"`
	VpcId		string	`qcloud_arg:"vpcId"`
	SubnetId	string	`qcloud_arg:"subnetId"`
	OsName		string  `qcloud_arg:"osName"`
	BandwidthType  string `qcloud_arg:"bandwidthType"`
	RootSize	int 	`qcloud_arg:"rootSize"`
	ClusterCIDR  string  `qcloud_arg:"clusterCIDR"`
	IsVpcGateway  int   `qcloud_arg:"isVpcGateway"`
	WanIp		int `qcloud_arg:"wanIp"`
	Password    string  `qcloud_arg:"password"`
}



type CreateClusterResponse struct{
	code		int			`json:"code"`
	codeDesc 	string		`json:"codeDesc"`
	message		string		`json:"message"`
	requestId	int			`json:"requestId"`
	clusterId	string		`json:"clusterId"`
}

type CreateEmptyClusterArgs struct {
	ClusterName string	`qcloud_arg:"clusterName"`
	ClusterCIDR  string  `qcloud_arg:"clusterCIDR"`
	ClusterVersion string `qcloud_arg:"clusterVersion"`
	VpcId		string	`qcloud_arg:"vpcId"`
	Region      string  `qcloud_arg:"region"`
}


type CreateEmptyClusterResponse struct {
	Code		int			`json:"code"`
	Message		string		`json:"message"`
	CodeDesc 	string		`json:"codeDesc"`
	Data struct {
		RequestId  int	`json:"requestId"`
		ClusterId	string		`json:"clusterId"`
	} `json:"data"`
}
type DeleteClusterArgs struct {
	ClusterId string	`qcloud_arg:"clusterId"`
}

type DeleteClusterResponse struct {
	code		int			`json:"code"`
	codeDesc 	string		`json:"codeDesc"`
	message		string		`json:"message"`
	requestId	int			`json:"requestId"`
}

func (client *Client) DeleteCluster (args *DeleteClusterArgs) (*DeleteClusterResponse, error){
	realRsp := &DeleteClusterResponse{}
	err := client.Invoke("DeleteCluster", args, realRsp)
	if err != nil {
		return &DeleteClusterResponse{}, err
	}
	return realRsp, nil
}


type DeleteClusterInstancesArgs struct {
	ClusterId string	`qcloud_arg:"clusterId"`
	InstanceIds []string `qcloud_arg:"instanceIds"`
}

type DeleteClusterInstancesResponse struct {
	code		int			`json:"code"`
	codeDesc 	string		`json:"codeDesc"`
	message		string		`json:"message"`
	requestId	int			`json:"requestId"`
}

func (client *Client) DeleteClusterInstances(args *DeleteClusterInstancesArgs) (*DeleteClusterInstancesResponse, error){
	realRsp := &DeleteClusterInstancesResponse{}
	err := client.Invoke("DeleteClusterInstances", args, realRsp)
	if err != nil {
		return &DeleteClusterInstancesResponse{}, err
	}
	return realRsp, nil
}

func (client *Client) CreateEmptyCluster (args *CreateEmptyClusterArgs) (*CreateEmptyClusterResponse, error){
	realRsp := &CreateEmptyClusterResponse{}
	err := client.Invoke("CreateEmptyCluster", args, realRsp)
	if err != nil {
		return &CreateEmptyClusterResponse{}, err
	}
	return realRsp, nil
}

//func (client *Client) CreateEmptyCluster (args *CreateEmptyClusterArgs) error{
//	realRsp := &CreateEmptyClusterResponse{}
//	err := client.Invoke("CreateEmptyCluster", args, realRsp)
//	if err != nil {
//		return &CreateEmptyClusterResponse{}, err
//	}
//	return realRsp, nil
//}

func (client *Client) DescribeCluster (args *DescribeClustersArgs) (*DescribeClustersResponse, error){
	realRsp := &DescribeClustersResponse{}
	err := client.Invoke("DescribeCluster", args, realRsp)
	if err != nil {
		return &DescribeClustersResponse{}, err
	}
	return realRsp, nil
}

func (client *Client) CreateCluster (args *CreateClusterArgs) (*CreateClusterResponse, error){
	realRsp := &CreateClusterResponse{}
	err := client.Invoke("CreateCluster", args, realRsp)
	if err != nil {
		return &CreateClusterResponse{}, err
	}
	return realRsp, nil
}

type AddClusterInstancesResponse struct{
	Code		int			`json:"code"`
	Message		string		`json:"message"`
	CodeDesc 	string		`json:"codeDesc"`
	Data struct {
		RequestId	int			`json:"requestId"`
		InstanceIds	 []string		`json:"instanceIds"`
	} `json:"data"`
}

type AddClusterInstancesArgs  struct {
	ClusterId string	`qcloud_arg:"clusterId"`
	ZoneId		string 	`qcloud_arg:"zoneId"`
	Cpu			int		`qcloud_arg:"cpu"`
	Mem			int		`qcloud_arg:"mem"`
	BandwidthType  string `qcloud_arg:"bandwidthType"`
	Bandwidth	int		`qcloud_arg:"bandwidth"`
	SubnetId	string	`qcloud_arg:"subnetId"`
	StorageSize	int		`qcloud_arg:"storageSize"`
	RootSize	int 	`qcloud_arg:"rootSize"`
	GoodsNum	int 	`qcloud_arg:"goodsNum"`
	Password    string  `qcloud_arg:"password"`
	IsVpcGateway  int   `qcloud_arg:"isVpcGateway"`
	WanIp		int 	`qcloud_arg:"wanIp"`
	OsName		string  `qcloud_arg:"osName"`
}




//func (client *Client) AddClusterInstances  (args *AddClusterInstancesArgs) string{
//	realRsp := &CreateEmptyClusterResponse{}
//	return client.Invoketest("AddClusterInstances", args, realRsp)
//	//realRsp := &AddClusterInstancesResponse{}
//	//err := client.Invoke("AddClusterInstances", args, realRsp)
//	//if err != nil {
//	//	return &AddClusterInstancesResponse{}, err
//	//}
//	//return realRsp, nil
//}

func (client *Client) AddClusterInstances  (args *AddClusterInstancesArgs) (*AddClusterInstancesResponse, error){
	realRsp := &AddClusterInstancesResponse{}
	err := client.Invoke("AddClusterInstances", args, realRsp)
	if err != nil {
		return &AddClusterInstancesResponse{}, err
	}
	return realRsp, nil
}

type DescribeClusterInstancesArgs struct {
	ClusterId  string `qcloud_arg:"clusterId"`
}

type DescribeClusterInstancesResponse struct{
	Code		int			`json:"code"`
	CodeDesc 	string		`json:"codeDesc"`
	Message		string		`json:"message"`
	Data struct {
		TotalCount		int			`json:"totalCount"`
		Node 			[]Nodes1			`json:"nodes"`
	} `json:"data"`
}

type Nodes1 struct {
	InstanceID			string   `json:"InstanceId"`
	InstanceName        string   `json:"InstanceName"`
	InstanceType        string   `json:"InstanceType"`
	kernelVersion       string	`json:"kernelVersion"`
	PodCidr			    string	`json:"podCidr"`
	CPU                 int      `json:"cpu"`
	Mem            	    int      `json:"mem"`
	WanIp				string   `json:"wanIp"`
	LanIp				string   `json:"lanIp"`
	OsImage				string   `json:"osImage"`
	IsNormal				int   `json:"isNormal"`
	CvmState				int   `json:"cvmState"`
	CvmPayMode				int   `json:"cvmPayMode"`
	NetworkPayMode				int   `json:"networkPayMode"`
	CreatedAt				string   `json:"createdAt"`
	InstanceCreateTime				string   `json:"instanceCreateTime"`
	InstanceDeadlineTime				string   `json:"instanceDeadlineTime"`
	ZoneId				int   `json:"zoneId"`
	Zone				string   `json:"zone"`
	AbnormalReason				string   `json:"abnormalReason"`
}

type Nodes struct {
	InstanceID			string   `json:"InstanceId"`
	InstanceName        string   `json:"InstanceName"`
	InstanceType        string   `json:"InstanceType"`
	kernelVersion       string	`json:"kernelVersion"`
	PodCidr			    string	`json:"podCidr"`
	CPU                 int      `json:"cpu"`
	Mem            	    int      `json:"mem"`
	WanIp				string   `json:"wanIp"`
	LanIp				string   `json:"lanIp"`
	OsImage				string   `json:"osImage"`
	IsNormal				int   `json:"isNormal"`
	CvmState				int   `json:"cvmState"`
	CvmPayMode				int   `json:"cvmPayMode"`
	NetworkPayMode				int   `json:"networkPayMode"`
	CreatedAt				string   `json:"createdAt"`
	InstanceCreateTime				string   `json:"instanceCreateTime"`
	InstanceDeadlineTime				string   `json:"instanceDeadlineTime"`
	ZoneId				int   `json:"zoneId"`
	Zone				string   `json:"zone"`
	AbnormalReason				string   `json:"abnormalReason"`
	Labels				map[string]string    `json:"labels"`
}


//func (client *Client) DescribeClusterInstances  (args *DescribeClusterInstancesArgs) string{
//	realRsp := &DescribeClusterInstancesResponse{}
//	return client.InvokeClusterInstanceTest("DescribeClusterInstances", args, realRsp)
//	//log.Println(s)
//	//if err != nil {
//	//	return "", err
//	//}
//	//return s, nil
//}

func (client *Client) DescribeClusterInstances  (args *DescribeClusterInstancesArgs) (*DescribeClusterInstancesResponse, error){
	realRsp := &DescribeClusterInstancesResponse{}
	err := client.Invoke("DescribeClusterInstances", args, realRsp)
	//log.Println(s)
	if err != nil {
		return &DescribeClusterInstancesResponse{}, err
	}
	return realRsp, nil
}