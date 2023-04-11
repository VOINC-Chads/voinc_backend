module "masters" {
  source   = "../masters"

  name              = "${var.uuid}-master"
  subnet_id         = var.subnet_id
  security_group_id = var.security_group_id
  gateway_id        = var.gateway_id
  image             = "master" # Need to make it so we use worker images too
}

# data "aws_instances" "ec2_instances" {
#   instance_ids = aws_instance.ec2_instances.*.id
# }

# data "aws_instance_public_ips" "ec2_instance_ips" {
#   for_each = { for idx, instance in aws_instance.ec2_instances.*.id : instance.id => instance }
#   ip_address = each.value.public_ip
# }

# locals {
#   ip_map = {for master in module.masters : master.name => master.server_public_ip }
#   local.ip_map.value["${var.uuid}-master"]
# }

module "workers" {
  count = var.n
  source   = "../workers"

  name              = "${var.uuid}-worker-${count.index}"
  subnet_id         = var.subnet_id
  security_group_id = var.security_group_id
  gateway_id        = var.gateway_id
  image             = "worker" # Need to create this still
  master_ip         = module.masters.server_public_ip
}

output "public-ip" {
  value = module.masters.*
}