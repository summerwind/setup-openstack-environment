package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
)

const (
	//osMetaDataAddr    = "169.254.169.254"
	osMetaDataAddr    = "127.0.0.1:8000"
	osMetaDataVersion = "2012-08-10"
)

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

func NewOpenStackMetaData(cdPath string) (*OpenStackMetaData, error) {
	var (
		buf []byte
		md  OpenStackMetaData
		err error
	)

	mdPath := filepath.Join("openstack", osMetaDataVersion, "meta_data.json")

	if cdPath == "" {
		mdURL := fmt.Sprintf("http://%s/%s", osMetaDataAddr, mdPath)
		resp, err := http.Get(mdURL)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		buf, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
	} else {
		p := filepath.Join(flagConfigDrive, mdPath)
		buf, err = ioutil.ReadFile(p)
		if err != nil {
			return nil, err
		}
	}

	err = json.Unmarshal(buf, &md)
	return &md, err
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

type OpenStackMetaDataFile struct {
	ContentPath string `json:"content_path"`
	Path        string `json:"path"`
}
