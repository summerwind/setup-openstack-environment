# setup-openstack-environment

Create an environment file with OpenStack metadata information.  
This command is inspired by [setup-network-environment](https://github.com/kelseyhightower/setup-network-environment).

## Usage

`setup-openstack-environment` command will get the metadata information from the OpenStack metadata server.

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

You can write the metadata information to a different file using the `-o` flag.

```
$ setup-network-environment -o /tmp/openstack-environment
```

With the `-c` flag, change the source of metadata infromation to the config drive.

```
$ setup-openstack-environment -c /mnt/config
```

## Build

```
$ make
```
