## Running the Google Cloud Platform templates to set up a Consul cluster

The platform variable defines the target OS, default is `ubuntu`.

Supported Machine Images:
- Ubuntu 14.04 (`ubuntu`)
- RHEL6 (`rhel6`)
- RHEL7 (`rhel7`)
- CentOS6 (`centos6`)
- CentOS7 (`centos7`)

For Google Cloud provider, set up your environment as outlined here: https://www.terraform.io/docs/providers/google/index.html

To set up a Ubuntu based cluster, replace `key_path` with actual value and run:


```shell
terraform apply -var 'key_path=/Users/xyz/consul.pem'
```

_or_

```shell
terraform apply -var 'key_path=/Users/xyz/consul.pem' -var 'platform=ubuntu'
```

To run RHEL6, run like below:

```shell
terraform apply -var 'key_path=/Users/xyz/consul.pem' -var 'platform=rhel6'
```

**Note:** For RHEL and CentOS based clusters, you need to have a [SSH key added](https://console.cloud.google.com/compute/metadata/sshKeys) for the user `root`.