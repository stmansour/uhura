// This code parses the file returned by 'aws ec2 run-instances --output json ...'
// It creates a data structure of type AwsNewInstance containing the parsed
// values of all the data and returns it to the caller.

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

type AwsNewInstance struct {
	Ownerid       string        `json:"OwnerId"`
	Reservationid string        `json:"ReservationId"`
	Groups        []interface{} `json:"Groups"`
	Instances     []struct {
		Monitoring struct {
			State string `json:"State"`
		} `json:"Monitoring"`
		Publicdnsname  string `json:"PublicDnsName"`
		Rootdevicetype string `json:"RootDeviceType"`
		State          struct {
			Code int    `json:"Code"`
			Name string `json:"Name"`
		} `json:"State"`
		Ebsoptimized          bool          `json:"EbsOptimized"`
		Launchtime            time.Time     `json:"LaunchTime"`
		Privateipaddress      string        `json:"PrivateIpAddress"`
		Productcodes          []interface{} `json:"ProductCodes"`
		Vpcid                 string        `json:"VpcId"`
		Statetransitionreason string        `json:"StateTransitionReason"`
		Instanceid            string        `json:"InstanceId"`
		Imageid               string        `json:"ImageId"`
		Privatednsname        string        `json:"PrivateDnsName"`
		Keyname               string        `json:"KeyName"`
		Securitygroups        []struct {
			Groupname string `json:"GroupName"`
			Groupid   string `json:"GroupId"`
		} `json:"SecurityGroups"`
		Clienttoken       string `json:"ClientToken"`
		Subnetid          string `json:"SubnetId"`
		Instancetype      string `json:"InstanceType"`
		Networkinterfaces []struct {
			Status             string `json:"Status"`
			Macaddress         string `json:"MacAddress"`
			Sourcedestcheck    bool   `json:"SourceDestCheck"`
			Vpcid              string `json:"VpcId"`
			Description        string `json:"Description"`
			Networkinterfaceid string `json:"NetworkInterfaceId"`
			Privateipaddresses []struct {
				Privatednsname   string `json:"PrivateDnsName"`
				Primary          bool   `json:"Primary"`
				Privateipaddress string `json:"PrivateIpAddress"`
			} `json:"PrivateIpAddresses"`
			Privatednsname string `json:"PrivateDnsName"`
			Attachment     struct {
				Status              string    `json:"Status"`
				Deviceindex         int       `json:"DeviceIndex"`
				Deleteontermination bool      `json:"DeleteOnTermination"`
				Attachmentid        string    `json:"AttachmentId"`
				Attachtime          time.Time `json:"AttachTime"`
			} `json:"Attachment"`
			Groups []struct {
				Groupname string `json:"GroupName"`
				Groupid   string `json:"GroupId"`
			} `json:"Groups"`
			Subnetid         string `json:"SubnetId"`
			Ownerid          string `json:"OwnerId"`
			Privateipaddress string `json:"PrivateIpAddress"`
		} `json:"NetworkInterfaces"`
		Sourcedestcheck bool `json:"SourceDestCheck"`
		Placement       struct {
			Tenancy          string `json:"Tenancy"`
			Groupname        string `json:"GroupName"`
			Availabilityzone string `json:"AvailabilityZone"`
		} `json:"Placement"`
		Hypervisor          string        `json:"Hypervisor"`
		Blockdevicemappings []interface{} `json:"BlockDeviceMappings"`
		Architecture        string        `json:"Architecture"`
		Statereason         struct {
			Message string `json:"Message"`
			Code    string `json:"Code"`
		} `json:"StateReason"`
		Rootdevicename     string `json:"RootDeviceName"`
		Virtualizationtype string `json:"VirtualizationType"`
		Amilaunchindex     int    `json:"AmiLaunchIndex"`
	} `json:"Instances"`
}

func LoadNewAwsInstanceInfo(fname string) *AwsNewInstance {
	file, e := ioutil.ReadFile(fname)
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}

	inst := new(AwsNewInstance)
	err := json.Unmarshal(file, inst)
	if nil != err {
		ulog("Could not unmarshal aws instance")
	}

	return inst
}
