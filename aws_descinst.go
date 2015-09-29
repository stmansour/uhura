// This code parses searches the return data from 'aws ec2 describe-instances --output json ...'
// for the supplied instance id.

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"
)

// AwsDescribeInstances is a type definition of the
// data we get back from the command 'aws ec2 describe-instances'.
// Uhura uses this data to determine the publicDNS of the instances
// it creates. I validated with AWS support... this is the best way
// to do it.
type AwsDescribeInstances struct {
	Reservations []struct {
		Ownerid       string        `json:"OwnerId"`
		Reservationid string        `json:"ReservationId"`
		Groups        []interface{} `json:"Groups"`
		Instances     []struct {
			Monitoring struct {
				State string `json:"State"`
			} `json:"Monitoring"`
			Publicdnsname string `json:"PublicDnsName"`
			Platform      string `json:"Platform"`
			State         struct {
				Code int    `json:"Code"`
				Name string `json:"Name"`
			} `json:"State"`
			Ebsoptimized          bool          `json:"EbsOptimized"`
			Launchtime            time.Time     `json:"LaunchTime"`
			Publicipaddress       string        `json:"PublicIpAddress"`
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
				Status          string `json:"Status"`
				Macaddress      string `json:"MacAddress"`
				Sourcedestcheck bool   `json:"SourceDestCheck"`
				Vpcid           string `json:"VpcId"`
				Description     string `json:"Description"`
				Association     struct {
					Publicip      string `json:"PublicIp"`
					Publicdnsname string `json:"PublicDnsName"`
					Ipownerid     string `json:"IpOwnerId"`
				} `json:"Association"`
				Networkinterfaceid string `json:"NetworkInterfaceId"`
				Privateipaddresses []struct {
					Privatednsname string `json:"PrivateDnsName"`
					Association    struct {
						Publicip      string `json:"PublicIp"`
						Publicdnsname string `json:"PublicDnsName"`
						Ipownerid     string `json:"IpOwnerId"`
					} `json:"Association"`
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
			Hypervisor          string `json:"Hypervisor"`
			Blockdevicemappings []struct {
				Devicename string `json:"DeviceName"`
				Ebs        struct {
					Status              string    `json:"Status"`
					Deleteontermination bool      `json:"DeleteOnTermination"`
					Volumeid            string    `json:"VolumeId"`
					Attachtime          time.Time `json:"AttachTime"`
				} `json:"Ebs"`
			} `json:"BlockDeviceMappings"`
			Architecture       string `json:"Architecture"`
			Rootdevicetype     string `json:"RootDeviceType"`
			Rootdevicename     string `json:"RootDeviceName"`
			Virtualizationtype string `json:"VirtualizationType"`
			Tags               []struct {
				Value string `json:"Value"`
				Key   string `json:"Key"`
			} `json:"Tags"`
			Amilaunchindex int `json:"AmiLaunchIndex"`
		} `json:"Instances"`
	} `json:"Reservations"`
}

// AllAwsInstances contains all the information from 'aws ec2 describe-instances'.
// knowing the InstanceID of the instances we create, we use this call to determine
// their publicDNS address.
var AllAwsInstances AwsDescribeInstances

// ReadAllAwsInstances injests the json output 'aws ec2 describe-instances'
func ReadAllAwsInstances(fname string) {
	b, e := ioutil.ReadFile(fname)
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}
	e = json.Unmarshal(b, &AllAwsInstances)
	if e != nil {
		log.Fatal(e)
	}
}

func searchReservationsForPublicDNS(iid string) string {
	for i := 0; i < len(AllAwsInstances.Reservations); i++ {
		for j := 0; j < len(AllAwsInstances.Reservations[i].Instances); j++ {
			if iid == AllAwsInstances.Reservations[i].Instances[j].Instanceid {
				return AllAwsInstances.Reservations[i].Instances[j].Publicdnsname
			}
		}
	}
	return ""
}
