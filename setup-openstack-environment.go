package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
)

const (
	formatOpenStack = "openstack"
	formatEC2       = "ec2"
	envPrefix       = "OPENSTACK"
)

var (
	flagFormat      string
	flagOutput      string
	flagConfigDrive string
)

type MetaData interface {
	String() string
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
		md  MetaData
		err error
	)

	switch flagFormat {
	case formatOpenStack:
		md, err = NewOpenStackMetaData(flagConfigDrive)
		if err != nil {
			abort(fmt.Errorf("Unable to fetch metadata: %s", err))
		}
	case formatEC2:
		md, err = NewEC2MetaData(flagConfigDrive)
		if err != nil {
			abort(fmt.Errorf("Unable to fetch metadata: %s", err))
		}
	default:
		abort(fmt.Errorf("Invalid format: %s", flagFormat))
	}

	err = ioutil.WriteFile(flagOutput, []byte(md.String()), os.ModePerm)
	if err != nil {
		abort(err)
	}
}

func abort(err error) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	os.Exit(1)
}
