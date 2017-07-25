provider "digitalocean" {
  token = "${var.do_token}"
}

resource "digitalocean_droplet" "consul" {
  ssh_keys           = ["${var.ssh_key_ID}"]
  image              = "${var.ubuntu}"
  region             = "${var.region}"
  size               = "2gb"
  private_networking = true
  name               = "consul${count.index + 1}"
  count              = "${var.num_instances}"

  connection {
    type        = "ssh"
    private_key = "${file("${var.key_path}")}"
    user        = "root"
    timeout     = "2m"
  }

  provisioner "file" {
    source      = "${path.module}/../shared/scripts/debian_upstart.conf"
    destination = "/tmp/upstart.conf"
  }

  provisioner "remote-exec" {
    inline = [
      "echo ${var.num_instances} > /tmp/consul-server-count",
      "echo ${digitalocean_droplet.consul.0.ipv4_address} > /tmp/consul-server-addr",
    ]
  }

  provisioner "remote-exec" {
    scripts = [
      "${path.module}/../shared/scripts/install.sh",
      "${path.module}/../shared/scripts/service.sh",
      "${path.module}/../shared/scripts/ip_tables.sh",
    ]
  }
}
