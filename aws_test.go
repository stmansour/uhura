package main

import (
	"testing"
)

// Read in json and validate that we can identify the InstanceId
func TestNewInstParse(t *testing.T) {
	nid := "i-1cd365bf"
	newinst := LoadNewAwsInstanceInfo("./test/publicdns/new-linux-jenkins-instance.json")
	if nid != newinst.Instances[0].Instanceid {
		t.Errorf("New Instanceid is wrong, expected %s, found %s", nid, newinst.Instances[0].Instanceid)
	}
}

func TestLoadAwsInstances(t *testing.T) {
	pubDNS := SearchReservationsForPublicDNS("./test/publicdns/describe-instances.json", "i-1cd365bf")
	if "ec2-54-208-20-125.compute-1.amazonaws.com" != pubDNS {
		t.Errorf("PubDNS: expected ec2-54-208-20-125.compute-1.amazonaws.com, got %s", pubDNS)
	}

	pubDNS = SearchReservationsForPublicDNS("./test/publicdns/describe-instances.json", "total-garbage")
	if "" != pubDNS {
		t.Errorf("PubDNS: expected empty string, got %s", pubDNS)
	}
}
