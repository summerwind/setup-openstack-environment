package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	ec2MetaDataAddr    = "169.254.169.254"
	ec2MetaDataVersion = "2009-04-04"
)

type EC2MetaData struct {
	AMIID              string                          `json:"ami-id"`
	AMILaunchIndex     int                             `json:"ami-launch-index"`
	AMIManifestPath    string                          `json:"ami-manifest-path"`
	BlockDeviceMapping map[string]string               `json:"block-device-mapping"`
	Hostname           string                          `json:"hostname"`
	InstanceAction     string                          `json:"instance-action"`
	InstanceID         string                          `json:"instance-id"`
	InstanceType       string                          `json:"instance-type"`
	KernelID           string                          `json:"kernel-id"`
	LocalHostname      string                          `json:"local-hostname"`
	LocalIPv4          string                          `json:"local-ipv4"`
	Placement          map[string]string               `json:"placement"`
	PublicHostname     string                          `json:"public-hostname"`
	PublicIPv4         string                          `json:"public-ipv4"`
	PublicKeys         map[string]EC2MetaDataPublicKey `json:"public-keys"`
	RAMDiskID          string                          `json:"ramdisk-id"`
	ReservationID      string                          `json:"reservation-id"`
	SecurityGroups     []string                        `json:"security-groups"`
}

func NewEC2MetaData(cdPath string) (*EC2MetaData, error) {
	var (
		buf []byte
		md  EC2MetaData
		err error
	)

	if cdPath != "" {
		p := filepath.Join(flagConfigDrive, "ec2", ec2MetaDataVersion, "meta-data.json")
		buf, err = ioutil.ReadFile(p)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(buf, &md)
		return &md, err
	}

	md.BlockDeviceMapping = map[string]string{}
	md.Placement = map[string]string{}
	md.PublicKeys = map[string]EC2MetaDataPublicKey{}

	baseURL := fmt.Sprintf("http://%s/ec2/%s/meta-data", ec2MetaDataAddr, ec2MetaDataVersion)

	md.AMIID, err = fetch(baseURL, "ami-id")
	if err != nil {
		return nil, err
	}

	launchIndex, err := fetch(baseURL, "ami-launch-index")
	if err != nil {
		return nil, err
	}

	md.AMILaunchIndex, err = strconv.Atoi(launchIndex)
	if err != nil {
		return nil, err
	}

	md.AMIManifestPath, err = fetch(baseURL, "ami-manifest-path")
	if err != nil {
		return nil, err
	}

	blockDevices, err := fetch(baseURL, "block-device-mapping")
	if err != nil {
		return nil, err
	}

	blockDeviceKeys := strings.Split(blockDevices, "\n")
	for _, key := range blockDeviceKeys {
		val, err := fetch(baseURL, fmt.Sprintf("block-device-mapping/%s", key))
		if err != nil {
			return nil, err
		}
		md.BlockDeviceMapping[key] = val
	}

	md.Hostname, err = fetch(baseURL, "hostname")
	if err != nil {
		return nil, err
	}

	md.InstanceAction, err = fetch(baseURL, "instance-action")
	if err != nil {
		return nil, err
	}

	md.InstanceID, err = fetch(baseURL, "instance-id")
	if err != nil {
		return nil, err
	}

	md.InstanceType, err = fetch(baseURL, "instance-type")
	if err != nil {
		return nil, err
	}

	md.LocalHostname, err = fetch(baseURL, "local-hostname")
	if err != nil {
		return nil, err
	}

	md.LocalIPv4, err = fetch(baseURL, "local-ipv4")
	if err != nil {
		return nil, err
	}

	placement, err := fetch(baseURL, "placement")
	if err != nil {
		return nil, err
	}

	placementKeys := strings.Split(placement, "\n")
	for _, key := range placementKeys {
		val, err := fetch(baseURL, fmt.Sprintf("placement/%s", key))
		if err != nil {
			return nil, err
		}
		md.Placement[key] = val
	}

	md.PublicHostname, err = fetch(baseURL, "public-hostname")
	if err != nil {
		return nil, err
	}

	md.PublicIPv4, err = fetch(baseURL, "public-ipv4")
	if err != nil {
		return nil, err
	}

	publicKeys, err := fetch(baseURL, "public-keys")
	if err != nil {
		return nil, err
	}

	pkKeys := strings.Split(publicKeys, "\n")
	for _, key := range pkKeys {
		indexAndLabel := strings.Split(key, "=")
		if len(indexAndLabel) != 2 {
			return nil, fmt.Errorf("Invalid public key: %s", key)
		}

		val, err := fetch(baseURL, fmt.Sprintf("public-keys/%s/openssh-key", indexAndLabel[0]))
		if err != nil {
			return nil, err
		}

		md.PublicKeys[indexAndLabel[0]] = EC2MetaDataPublicKey{val}
	}

	md.RAMDiskID, err = fetch(baseURL, "ramdisk-id")
	if err != nil {
		return nil, err
	}

	md.ReservationID, err = fetch(baseURL, "reservation-id")
	if err != nil {
		return nil, err
	}

	securityGroups, err := fetch(baseURL, "security-groups")
	if err != nil {
		return nil, err
	}
	md.SecurityGroups = strings.Split(securityGroups, ",")

	return &md, err
}

func fetch(baseURL, key string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("%s/%s", baseURL, key))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return strings.TrimRight(string(buf), "\n"), nil
}

func (md EC2MetaData) String() string {
	var buf = bytes.NewBuffer(make([]byte, 0, 100))

	buf.WriteString(fmt.Sprintf("%s_AMI_ID=%s\n", envPrefix, md.AMIID))
	buf.WriteString(fmt.Sprintf("%s_AMI_LAUNCH_INDEX=%d\n", envPrefix, md.AMILaunchIndex))
	buf.WriteString(fmt.Sprintf("%s_AMI_MANIFEST_PATH=%s\n", envPrefix, md.AMIManifestPath))
	for key, val := range md.BlockDeviceMapping {
		key = strings.ToUpper(key)
		key = strings.Replace(key, "-", "_", -1)
		buf.WriteString(fmt.Sprintf("%s_BLOCK_DEVICE_MAPPING_%s=%s\n", envPrefix, key, strings.TrimSpace(val)))
	}
	buf.WriteString(fmt.Sprintf("%s_HOSTNAME=%s\n", envPrefix, md.Hostname))
	buf.WriteString(fmt.Sprintf("%s_INSTANCE_ACTION=%s\n", envPrefix, md.InstanceAction))
	buf.WriteString(fmt.Sprintf("%s_INSTANCE_ID=%s\n", envPrefix, md.InstanceID))
	buf.WriteString(fmt.Sprintf("%s_INSTANCE_TYPE=%s\n", envPrefix, md.InstanceType))
	buf.WriteString(fmt.Sprintf("%s_KERNEL_ID=%s\n", envPrefix, md.KernelID))
	buf.WriteString(fmt.Sprintf("%s_LOCAL_HOSTNAME=%s\n", envPrefix, md.LocalHostname))
	buf.WriteString(fmt.Sprintf("%s_LOCAL_IPV4=%s\n", envPrefix, md.LocalIPv4))
	for key, val := range md.Placement {
		key = strings.ToUpper(key)
		key = strings.Replace(key, "-", "_", -1)
		buf.WriteString(fmt.Sprintf("%s_PLACEMENT_%s=%s\n", envPrefix, key, strings.TrimSpace(val)))
	}
	buf.WriteString(fmt.Sprintf("%s_PUBLIC_HOSTNAME=%s\n", envPrefix, md.PublicHostname))
	buf.WriteString(fmt.Sprintf("%s_PUBLIC_IPV4=%s\n", envPrefix, md.PublicIPv4))
	for key, val := range md.PublicKeys {
		key = strings.ToUpper(key)
		key = strings.Replace(key, "-", "_", -1)
		buf.WriteString(fmt.Sprintf("%s_PUBLIC_KEYS_%s_OPENSSH_KEY=%s\n", envPrefix, key, strings.TrimSpace(val.OpenSSHKey)))
	}
	buf.WriteString(fmt.Sprintf("%s_RAM_DISK_ID=%s\n", envPrefix, md.RAMDiskID))
	buf.WriteString(fmt.Sprintf("%s_RESERVATION_ID=%s\n", envPrefix, md.ReservationID))
	for i, val := range md.SecurityGroups {
		buf.WriteString(fmt.Sprintf("%s_SECURITY_GROUPS_%d=%s\n", envPrefix, i, val))
	}

	return buf.String()
}

type EC2MetaDataPublicKey struct {
	OpenSSHKey string `json:"openssh-key"`
}
