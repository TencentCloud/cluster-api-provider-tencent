### Install tccli using 

https://intl.cloud.tencent.com/document/product/1013/33464

### To get images available 
```shell
tccli tke DescribeImages
{
    "TotalCount": 29,
    "ImageInstanceSet": [
        {
            "Alias": "CentOS 7.2 64bit",
            "OsName": "centos7.2x86_64",
            "ImageId": "img-rkiynh11",
            "OsCustomizeType": "GENERAL"
        },
        {
            "Alias": "CentOS 7.2 64bit GPU",
            "OsName": "centos7.2x86_64 GPU",
            "ImageId": "img-idtyzihp",
            "OsCustomizeType": "GENERAL"
        },
        {
            "Alias": "CentOS 7.2 64bit BMS",
            "OsName": "centos7.4 bm x86_64",
            "ImageId": "img-8toqc6s3",
            "OsCustomizeType": "GENERAL"
        },
        {
            "Alias": "CentOS 7.6 64bit TKE-Optimized",
            "OsName": "centos7.6.0_x64",
            "ImageId": "img-cgndmknl",
            "OsCustomizeType": "DOCKER_CUSTOMIZE"
        },
        {
            "Alias": "CentOS 7.6 64bit",
            "OsName": "centos7.6.0_x64",
            "ImageId": "img-9qabwvbn",
            "OsCustomizeType": "GENERAL"
        },
        {
            "Alias": "CentOS 7.6 64bit BMS",
            "OsName": "centos7.6.0_x64 BMS",
            "ImageId": "img-9qabwvbn",
            "OsCustomizeType": "GENERAL"
        },
        {
            "Alias": "CentOS 7.6 64bit GPU",
            "OsName": "centos7.6.0_x64 GPU",
            "ImageId": "img-cgndmknl",
            "OsCustomizeType": "DOCKER_CUSTOMIZE"
        },
        {
            "Alias": "CentOS 7.6 64bit GPU",
            "OsName": "centos7.6.0_x64 GPU",
            "ImageId": "img-9qabwvbn",
            "OsCustomizeType": "GENERAL"
        },
        {
            "Alias": "CentOS 7.8 64bit",
            "OsName": "centos7.8.0_x64",
            "ImageId": "img-3la7wgnt",
            "OsCustomizeType": "GENERAL"
        },
        {
            "Alias": "CentOS 7.8 64bit BMS",
            "OsName": "centos7.8.0_x64 BMS",
            "ImageId": "img-3la7wgnt",
            "OsCustomizeType": "GENERAL"
        },
        {
            "Alias": "CentOS 7.8 64bit GPU",
            "OsName": "centos7.8.0_x64 GPU",
            "ImageId": "img-3la7wgnt",
            "OsCustomizeType": "GENERAL"
        },
        {
            "Alias": "CentOS 8.0 64bit",
            "OsName": "centos8.0x86_64",
            "ImageId": "img-25szkc8t",
            "OsCustomizeType": "GENERAL"
        },
        {
            "Alias": "CentOS 8.0 64bit BMS",
            "OsName": "centos8.0x86_64 BMS",
            "ImageId": "img-25szkc8t",
            "OsCustomizeType": "GENERAL"
        },
        {
            "Alias": "CentOS 8.0 64bit GPU",
            "OsName": "centos8.0x86_64 GPU",
            "ImageId": "img-25szkc8t",
            "OsCustomizeType": "GENERAL"
        },
        {
            "Alias": "CentOS 8.2(arm64)",
            "OsName": "centos8.2arm_64",
            "ImageId": "img-n74hgdxx",
            "OsCustomizeType": "GENERAL"
        },
        {
            "Alias": "CentOS 8.2(arm64) BMS",
            "OsName": "centos8.2arm_64 BMS",
            "ImageId": "img-n74hgdxx",
            "OsCustomizeType": "GENERAL"
        },
        {
            "Alias": "CentOS 8.2(arm64) GPU",
            "OsName": "centos8.2arm_64 GPU",
            "ImageId": "img-n74hgdxx",
            "OsCustomizeType": "GENERAL"
        },
        {
            "Alias": "CentOS 7.4(arm64)",
            "OsName": "centos7.4arm_64",
            "ImageId": "img-k4xgkxa5",
            "OsCustomizeType": "GENERAL"
        },
        {
            "Alias": "CentOS 7.4(arm64) GPU",
            "OsName": "centos7.4arm_64 GPU",
            "ImageId": "img-k4xgkxa5",
            "OsCustomizeType": "GENERAL"
        },
        {
            "Alias": "Ubuntu Server 16.04.1 LTS 64bit",
            "OsName": "ubuntu16.04.1 LTSx86_64",
            "ImageId": "img-4wpaazux",
            "OsCustomizeType": "GENERAL"
        },
        {
            "Alias": "Ubuntu Server 16.04.1 LTS 64bit GPU",
            "OsName": "ubuntu16.04.1 LTSx86_64 GPU",
            "ImageId": "img-h1qos5y5",
            "OsCustomizeType": "GENERAL"
        },
        {
            "Alias": "Ubuntu Server 18.04.1 LTS 64bit TKE-Optimized",
            "OsName": "ubuntu18.04.1x86_64",
            "ImageId": "img-8f4a3ri5",
            "OsCustomizeType": "DOCKER_CUSTOMIZE"
        },
        {
            "Alias": "Ubuntu Server 18.04.1 LTS 64bit",
            "OsName": "ubuntu18.04.1x86_64",
            "ImageId": "img-pi0ii46r",
            "OsCustomizeType": "GENERAL"
        },
        {
            "Alias": "Ubuntu Server 18.04.1 LTS 64bit GPU",
            "OsName": "ubuntu18.04.1x86_64 GPU",
            "ImageId": "img-8f4a3ri5",
            "OsCustomizeType": "DOCKER_CUSTOMIZE"
        },
        {
            "Alias": "Ubuntu Server 18.04.1 LTS 64bit GPU",
            "OsName": "ubuntu18.04.1x86_64 GPU",
            "ImageId": "img-pi0ii46r",
            "OsCustomizeType": "GENERAL"
        },
        {
            "Alias": "Ubuntu Server 20.04.1 LTS 64bit",
            "OsName": "ubuntu20.04x86_64",
            "ImageId": "img-22trbn9x",
            "OsCustomizeType": "GENERAL"
        },
        {
            "Alias": "Ubuntu Server 20.04.1 LTS 64bit GPU",
            "OsName": "ubuntu20.04x86_64 GPU",
            "ImageId": "img-22trbn9x",
            "OsCustomizeType": "GENERAL"
        },
        {
            "Alias": "CentOS 7.6 64位 + SG1-pv1.3",
            "OsName": "centos7.6.0_x64_sg1-pv1.3",
            "ImageId": "img-7ilszol5",
            "OsCustomizeType": "GENERAL"
        },
        {
            "Alias": "Ubuntu Server 20.04 LTS 64位 (Tencent Kernel 4)",
            "OsName": "ubuntu20.04(tkernel4)x86_64",
            "ImageId": "img-3m2r69mh",
            "OsCustomizeType": "GENERAL"
        }
    ],
    "RequestId": "73e56a29-00ed-4259-a234-1b76bf3c0145"
}
```