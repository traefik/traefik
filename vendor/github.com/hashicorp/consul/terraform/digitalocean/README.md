# Requirements

* Terraform installed
* Digital Ocean account with API key
* SSH key uploaded to Digital Ocean

### Variables
Populate terraform.tfvars as follows (or execute with arguments as shown in Usage)

    key_path = "~/.ssh/id_rsa"
    do_token = "ASDFQWERTYDERP"
    num_instances = "3"
    ssh_key_ID = "my_ssh_keyID_in_digital_ocean"
    region = "desired DO region"

# Usage

    terraform plan                      \
      -var 'key_path=~/.ssh/id_rsa'     \
      -var 'do_token=ASDFQWERTYDERP'    \
      -var 'num_instances=3'            \
      -var 'ssh_key_ID=86:75:30:99:88:88:AA:FF:DD' \
      -var 'region=tor1'

    terraform apply                     \
      -var 'key_path=~/.ssh/id_rsa'     \
      -var 'do_token=ASDFQWERTYDERP'    \
      -var 'num_instances=3'            \
      -var 'ssh_key_ID=86:75:30:99:88:88:AA:FF:DD' \
      -var 'region=tor1'
