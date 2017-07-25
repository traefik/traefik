variable "do_token" {}

variable "key_path" {}

variable "ssh_key_ID" {}

variable "region" {}

variable "num_instances" {}

# Default OS

variable "ubuntu" {
  description = "Default LTS"
  default     = "ubuntu-14-04-x64"
}

variable "centos" {
  description = "Default Centos"
  default     = "centos-72-x64"
}

variable "coreos" {
  description = "Defaut Coreos"
  default     = "coreos-899.17.0"
}
