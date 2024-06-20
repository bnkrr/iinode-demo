package main

import (
	"os"
	"strings"
)

func dbusId() (string, error) {
	id, err := os.ReadFile("/var/lib/dbus/machine-id")
	if err != nil {
		id, err = os.ReadFile("/etc/machine-id")
	}
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(strings.Trim(string(id), "\n")), nil
}

type NodeMetadata struct {
	PublicIp string
	Id       string
}

func (n *NodeMetadata) GenerateId() {
	id, err := dbusId()
	if err != nil {
		id = "default"
	}
	n.Id = id
}

func (n *NodeMetadata) GetPublicIp() {
	n.PublicIp = "1.2.3.5"
}

func (n *NodeMetadata) Refresh() {
	n.GenerateId()
	n.GetPublicIp()
}
