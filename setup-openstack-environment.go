package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const (
	envPrefix   = "OPENSTACK"
	osFilePath  = "openstack/2012-08-10/meta_data.json"
	ec2FilePath = "ec2/2009-04-04/meta-data.json"

	formatOpenStack = "openstack"
	formatEC2       = "ec2"
)

var (
	flagFormat      string
	flagOutput      string
	flagConfigDrive string
)

type MetaData interface {
	String() string
}

type OpenStackMetaData struct {
	AvailabilityZone string                  `json:"availability_zone"`
	Files            []OpenStackMetaDataFile `json:"files"`
	Hostname         string                  `json:"hostname"`
	LaunchIndex      int                     `json:"launch_index"`
	Name             string                  `json:"Name"`
	Meta             map[string]string       `json:"meta"`
	PublicKeys       map[string]string       `json:"public_keys"`
	UUID             string                  `json:"uuid"`
}

type OpenStackMetaDataFile struct {
	ContentPath string `json:"content_path"`
	Path        string `json:"path"`
}

func (md OpenStackMetaData) String() string {
	var buf = bytes.NewBuffer(make([]byte, 0, 100))

	buf.WriteString(fmt.Sprintf("%s_AVAILABILITY_ZONE=%s\n", envPrefix, md.AvailabilityZone))
	for i, f := range md.Files {
		buf.WriteString(fmt.Sprintf("%s_FILES_%d_CONTENT_PATH=%s\n", envPrefix, i, f.ContentPath))
		buf.WriteString(fmt.Sprintf("%s_FILES_%d_PATH=%s\n", envPrefix, i, f.Path))
	}
	buf.WriteString(fmt.Sprintf("%s_HOSTNAME=%s\n", envPrefix, md.Hostname))
	buf.WriteString(fmt.Sprintf("%s_LAUNCH_INDEX=%d\n", envPrefix, md.LaunchIndex))
	buf.WriteString(fmt.Sprintf("%s_NAME=%s\n", envPrefix, md.Name))
	for key, val := range md.Meta {
		buf.WriteString(fmt.Sprintf("%s_META_%s=%s\n", envPrefix, strings.ToUpper(key), val))
	}
	for key, val := range md.PublicKeys {
		buf.WriteString(fmt.Sprintf("%s_PUBLIC_KEYS_%s=%s\n", envPrefix, strings.ToUpper(key), strings.TrimSpace(val)))
	}
	buf.WriteString(fmt.Sprintf("%s_UUID=%s\n", envPrefix, md.UUID))

	return buf.String()
}

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

type EC2MetaDataPublicKey struct {
	OpenSSHKey string `json:"openssh-key"`
}

func (md EC2MetaData) String() string {
	var buf = bytes.NewBuffer(make([]byte, 0, 100))

	buf.WriteString(fmt.Sprintf("%s_AMI_ID=%s\n", envPrefix, md.AMIID))
	buf.WriteString(fmt.Sprintf("%s_AMI_LAUNCH_INDEX=%d\n", envPrefix, md.AMILaunchIndex))
	buf.WriteString(fmt.Sprintf("%s_AMI_MANIFEST_PATH=%s\n", envPrefix, md.AMIManifestPath))
	for key, val := range md.BlockDeviceMapping {
		buf.WriteString(fmt.Sprintf("%s_BLOCK_DEVICE_MAPPING_%s=%s\n", envPrefix, strings.ToUpper(key), strings.TrimSpace(val)))
	}
	buf.WriteString(fmt.Sprintf("%s_HOSTNAME=%s\n", envPrefix, md.Hostname))
	buf.WriteString(fmt.Sprintf("%s_INSTANCE_ACTION=%s\n", envPrefix, md.InstanceAction))
	buf.WriteString(fmt.Sprintf("%s_INSTANCE_ID=%s\n", envPrefix, md.InstanceID))
	buf.WriteString(fmt.Sprintf("%s_INSTANCE_TYPE=%s\n", envPrefix, md.InstanceType))
	buf.WriteString(fmt.Sprintf("%s_KERNEL_ID=%s\n", envPrefix, md.KernelID))
	buf.WriteString(fmt.Sprintf("%s_LOCAL_HOSTNAME=%s\n", envPrefix, md.LocalHostname))
	buf.WriteString(fmt.Sprintf("%s_LOCAL_IPV4=%s\n", envPrefix, md.LocalIPv4))
	for key, val := range md.Placement {
		buf.WriteString(fmt.Sprintf("%s_PLACEMENT_%s=%s\n", envPrefix, strings.ToUpper(key), strings.TrimSpace(val)))
	}
	buf.WriteString(fmt.Sprintf("%s_PUBLIC_HOSTNAME=%s\n", envPrefix, md.PublicHostname))
	buf.WriteString(fmt.Sprintf("%s_PUBLIC_IPV4=%s\n", envPrefix, md.PublicIPv4))
	for key, val := range md.PublicKeys {
		buf.WriteString(fmt.Sprintf("%s_PUBLIC_KEYS_%s_OPENSSH_KEY=%s\n", envPrefix, strings.ToUpper(key), strings.TrimSpace(val.OpenSSHKey)))
	}
	buf.WriteString(fmt.Sprintf("%s_RAM_DISK_ID=%s\n", envPrefix, md.RAMDiskID))
	buf.WriteString(fmt.Sprintf("%s_RESERVATION_ID=%s\n", envPrefix, md.ReservationID))
	for i, val := range md.SecurityGroups {
		buf.WriteString(fmt.Sprintf("%s_SECURITY_GROUPS_%d=%s\n", envPrefix, i, val))
	}

	return buf.String()
}

func abort(err error) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	os.Exit(1)
}

func main() {
	var cmd = &cobra.Command{
		Use:   "setup-openstack-environment",
		Short: "Create an environment file with openstack information.",
		Run:   run,
	}

	cmd.Flags().StringVarP(&flagFormat, "format", "f", "openstack", "Meta data format (\"openstack\" or \"ec2\")")
	cmd.Flags().StringVarP(&flagOutput, "output", "o", "/etc/openstack-environment", "Path of output file")
	cmd.Flags().StringVarP(&flagConfigDrive, "config-drive", "c", "", "Path of config drive")

	err := cmd.Execute()
	if err != nil {
		abort(err)
	}
}

func run(cmd *cobra.Command, args []string) {
	var (
		buf []byte
		err error
	)

	if flagFormat != formatOpenStack && flagFormat != formatEC2 {
		abort(fmt.Errorf("Invalid format: %s", flagFormat))
	}

	p := getPath(flagFormat)

	if flagConfigDrive != "" {
		p = filepath.Join(flagConfigDrive, p)

		buf, err = ioutil.ReadFile(p)
		if err != nil {
			abort(err)
		}
	} else {
		resp, err := http.Get(fmt.Sprintf("http://169.254.169.254/%s", p))
		if err != nil {
			abort(err)
		}
		defer resp.Body.Close()

		buf, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			abort(err)
		}
	}

	md, err := decodeMetaData(flagFormat, buf)
	if err != nil {
		abort(err)
	}

	err = ioutil.WriteFile(flagOutput, []byte(md.String()), os.ModePerm)
	if err != nil {
		abort(err)
	}
}

func getPath(format string) string {
	var p string

	switch format {
	case formatOpenStack:
		p = osFilePath
	case formatEC2:
		p = ec2FilePath
	}

	return p
}

func decodeMetaData(format string, buf []byte) (MetaData, error) {
	switch format {
	case formatOpenStack:
		var md OpenStackMetaData
		err := json.Unmarshal(buf, &md)
		return md, err
	case formatEC2:
		var md EC2MetaData
		err := json.Unmarshal(buf, &md)
		return md, err
	}

	return nil, fmt.Errorf("Invalid format: %s", format)
}
