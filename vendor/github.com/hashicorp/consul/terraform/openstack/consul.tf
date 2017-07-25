provider "openstack" {
    user_name  = "${var.username}"
    tenant_name = "${var.tenant_name}"
    password  = "${var.password}"
    auth_url  = "${var.auth_url}"
}

resource "openstack_compute_keypair_v2" "consul_keypair" {
  name = "consul-keypair"
  region = "${var.region}"
  public_key = "${var.public_key}"
}

resource "openstack_compute_floatingip_v2" "consul_ip" {
  region = "${var.region}"
  pool = "${lookup(var.pub_net_id, var.region)}"
  count = "${var.servers}"
}

resource "openstack_compute_instance_v2" "consul_node" {
  name = "consul-node-${count.index}"
  region = "${var.region}"
  image_id = "${lookup(var.image, var.region)}"
  flavor_id = "${lookup(var.flavor, var.region)}"
  floating_ip = "${element(openstack_compute_floatingip_v2.consul_ip.*.address,count.index)}"
  key_pair = "consul-keypair"
  count = "${var.servers}"

    connection {
        user = "${var.user_login}"
        key_file = "${var.key_file_path}"
        timeout = "1m"
    }

    provisioner "file" {
        source = "${path.module}/scripts/upstart.conf"
        destination = "/tmp/upstart.conf"
    }

    provisioner "file" {
        source = "${path.module}/scripts/upstart-join.conf"
        destination = "/tmp/upstart-join.conf"
    }

    provisioner "remote-exec" {
        inline = [
            "echo ${var.servers} > /tmp/consul-server-count",
            "echo ${count.index} > /tmp/consul-server-index",
            "echo ${openstack_compute_instance_v2.consul_node.0.network.0.fixed_ip_v4} > /tmp/consul-server-addr",
        ]
    }

    provisioner "remote-exec" {
        scripts = [
            "${path.module}/scripts/install.sh",
            "${path.module}/scripts/server.sh",
            "${path.module}/scripts/service.sh",
        ]
    }
}
