package main

import (
	"testing"
)

// Read in json and validate that we can identify the InstanceId
func TestNewInstParse(t *testing.T) {
	nid := "i-3c601794"
	newinst := AWSLoadNewInstanceInfo("./test/publicdns/new-linux-jenkins-instance.json")
	if nid != newinst.Instances[0].Instanceid {
		t.Errorf("New Instanceid is wrong, expected %s, found %s", nid, newinst.Instances[0].Instanceid)
	}
}

func TestLoadAwsInstances(t *testing.T) {
	ReadAllAwsInstances("./test/publicdns/describe-instances.json")

	pubDNS := SearchReservationsForPublicDNS("i-3c601794")
	if "ec2-52-6-164-191.compute-1.amazonaws.com" != pubDNS {
		t.Errorf("PubDNS: expected ec2-52-6-164-191.compute-1.amazonaws.com, got %s", pubDNS)
	}

	pubDNS = SearchReservationsForPublicDNS("total-garbage")
	if "" != pubDNS {
		t.Errorf("PubDNS: expected empty string, got %s", pubDNS)
	}
}
