# setup-openstack-environment

Create an environment file with OpenStack metadata information.  
This command is inspired by [setup-network-environment](https://github.com/kelseyhightower/setup-network-environment).

## Install

### Just want the binary?

Go to the [releases page](https://github.com/summerwind/setup-openstack-environment/releases), find the version you want, and download the tarball file.

### Building binary yourself

To build the binary you need to install [Go](https://golang.org/), [dep](https://github.com/golang/dep) and [task](https://github.com/go-task/task).

```
$ task vendor
$ task build
```

## Usage

`setup-openstack-environment` command will get the metadata information from the OpenStack metadata.

```
$ setup-openstack-environment
```

The metadata information will be written to `/etc/openstack-environment` by default.

```
$ cat /etc/openstack-environment
OPENSTACK_AVAILABILITY_ZONE=nova
OPENSTACK_HOSTNAME=test.novalocal
OPENSTACK_LAUNCH_INDEX=0
OPENSTACK_NAME=test
```

You can specify the output file path using the `-o` flag.

```
$ setup-network-environment -o /tmp/openstack-environment
```

With the `-c` flag, change the source of metadata infromation to the config drive.

```
$ setup-openstack-environment -c /mnt/config
```

By using the `-f` flag, the command will use metadata in EC2 format.

```
$ setup-openstack-environment -f ec2
```

