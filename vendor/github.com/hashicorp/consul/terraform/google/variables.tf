variable "platform" {
    default = "ubuntu"
    description = "The OS Platform"
}

variable "user" {
    default = {
        ubuntu  = "ubuntu"
        rhel6   = "root"
        rhel7   = "root"
        centos6 = "root"
        centos7 = "root"
    }
}

variable "machine_image" {
    default = {
        ubuntu  = "ubuntu-os-cloud/ubuntu-1404-trusty-v20160314"
        rhel6   = "rhel-cloud/rhel-6-v20160303"
        rhel7   = "rhel-cloud/rhel-7-v20160303"
        centos6 = "centos-cloud/centos-6-v20160301"
        centos7 = "centos-cloud/centos-7-v20160301"
    }
}

variable "service_conf" {
    default = {
        ubuntu  = "debian_upstart.conf"
        rhel6   = "rhel_upstart.conf"
        rhel7   = "rhel_consul.service"
        centos6 = "rhel_upstart.conf"
        centos7 = "rhel_consul.service"
    }
}
variable "service_conf_dest" {
    default = {
        ubuntu  = "upstart.conf"
        rhel6   = "upstart.conf"
        rhel7   = "consul.service"
        centos6 = "upstart.conf"
        centos7 = "consul.service"
    }
}

variable "key_path" {
    description = "Path to the private key used to access the cloud servers"
}

variable "region" {
    default     = "us-central1"
    description = "The region of Google Cloud where to launch the cluster"
}

variable "region_zone" {
    default     = "us-central1-f"
    description = "The zone of Google Cloud in which to launch the cluster"
}

variable "servers" {
    default     = "3"
    description = "The number of Consul servers to launch"
}

variable "machine_type" {
    default     = "f1-micro"
    description = "Google Cloud Compute machine type"
}

variable "tag_name" {
    default     = "consul"
    description = "Name tag for the servers"
}
