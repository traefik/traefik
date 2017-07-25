output "first_consule_node_address" {
  value = "${digitalocean_droplet.consul.0.ipv4_address}"
}

output "all_addresses" {
  value = ["${digitalocean_droplet.consul.*.ipv4_address}"]
}
