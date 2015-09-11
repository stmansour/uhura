package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

type AwsRes struct {
	Ownerid       string        `json:"OwnerId"`
	Reservationid string        `json:"ReservationId"`
	Groups        []interface{} `json:"Groups"`
	Instances     []struct {
		Monitoring struct {
			State string `json:"State"`
		} `json:"Monitoring"`
		Publicdnsname string `json:"PublicDnsName"`
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
}

//type AwsResList map[string]AwsRes

// Return the publicDNS name for the supplied instance
func SearchReservationsForPublicDNS(fname, instid string) string {
	file, e := ioutil.ReadFile(fname)
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}

	d := json.NewDecoder(strings.NewReader(string(file)))

	// read the open bracket...
	_, err := d.Token()
	if err != nil {
		panic(err)
	}
	var s *AwsRes
	for d.More() { // while the data stream contains values
		s = new(AwsRes)
		err := d.Decode(&s) // decode a struct
		if err != nil {
			panic(err)
		}
		if instid == s.Instances[0].Instanceid {
			return s.Instances[0].Publicdnsname
		}
	}
	return ""
}
