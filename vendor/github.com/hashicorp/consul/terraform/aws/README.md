## Running the aws templates to set up a consul cluster

The platform variable defines the target OS (which in turn controls whether we install the Consul service via `systemd` or `upstart`).  Options include:

- `ubuntu` (default)
- `rhel6`
- `rhel7`
- `centos6`
- `centos7`


For AWS provider, set up your AWS environment as outlined in https://www.terraform.io/docs/providers/aws/index.html

To set up ubuntu based, run the following command, taking care to replace `key_name` and `key_path` with actual values:

`terraform apply -var 'key_name=consul' -var 'key_path=/Users/xyz/consul.pem'`

or

`terraform apply -var 'key_name=consul' -var 'key_path=/Users/xyz/consul.pem' -var 'platform=ubuntu'`

For CentOS7:

`terraform apply -var 'key_name=consul' -var 'key_path=/Users/xyz/consul.pem' -var 'platform=centos7'`

For centos6 platform, for the default AMI, you need to accept the AWS market place terms and conditions. When you launch first time, you will get an error with an URL to accept the terms and conditions.
