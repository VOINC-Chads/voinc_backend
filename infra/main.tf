locals {
  infra = jsondecode(file("infra.json"))
}

resource "aws_vpc" "vpcs" {
  count = length(local.infra)

  cidr_block = "10.${count.index}.0.0/16"
  tags = {
    Name = local.infra[count.index].uuid
  }
}

resource "aws_internet_gateway" "igws" {
  count = length(aws_vpc.vpcs)

  vpc_id = aws_vpc.vpcs[count.index].id
  tags = {
    Name = "${aws_vpc.vpcs[count.index].tags.Name}-igw"
  }
}

resource "aws_route_table" "route-tables" {
  count = length(aws_vpc.vpcs)

  vpc_id = aws_vpc.vpcs[count.index].id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.igws[count.index].id
  }

  route {
    ipv6_cidr_block = "::/0"
    gateway_id      = aws_internet_gateway.igws[count.index].id
  }

  tags = {
    Name = "${aws_vpc.vpcs[count.index].tags.Name}-route-table"
  }
}

resource "aws_subnet" "subnets" {
  count = length(aws_vpc.vpcs)

  vpc_id = aws_vpc.vpcs[count.index].id
  cidr_block = "10.${count.index}.1.0/24"
  availability_zone = "us-east-2c"
  tags = {
    Name = "${aws_vpc.vpcs[count.index].tags.Name}-subnet"
  }
}

resource "aws_route_table_association" "instance-route-table-assoc" {
  count = length(aws_vpc.vpcs)
  subnet_id      = aws_subnet.subnets[count.index].id
  route_table_id = aws_route_table.route-tables[count.index].id
}

resource "aws_security_group" "sgs" {
  count = length(local.infra)
  name        = "allow_web_traffic"
  description = "Allow Web inbound traffic"
  vpc_id      = aws_vpc.vpcs[count.index].id

  ingress {
    description = "HTTP"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "HTTP"
    from_port   = 8000
    to_port     = 8000
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "HTTPS"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${aws_vpc.vpcs[count.index].tags.Name}-allow_web"
  }
}

# module "masters" {
#   for_each = { for idx, info in jsondecode(file("infra.json")) : idx => item }
#   source   = "./masters"

#   name              = "${each.value.uuid}-master"
#   subnet_id         = aws_subnet.subnets[each.key].id
#   security_group_id = aws_security_group.sgs[each.key].id
#   gateway           = aws_internet_gateway.igws[each.key]
#   image             = "master" # Need to make it so we use worker images too
# }

# data "aws_instances" "ec2_instances" {
#   instance_ids = aws_instance.ec2_instances.*.id
# }

# data "aws_instance_public_ips" "ec2_instances" {
#   for_each = { for idx, instance in data.aws_instances.ec2_instances : instance.id => instance }
#   ip_address = each.value.public_ip
# }

# locals{
#   ip_map = { for id, ip in data.aws_instance_public_ips.ec2_instances : id => ip.ip_address }
# }

# module "workers" {
#   for_each = { for idx, info in jsondecode(file("infra.json")) : idx => item }
#   source   = "./workers"

#   workers = [
#     for i in range(each.value.n) :
#       {
#         name              = "${each.value.uuid}-worker-${i}"
#         subnet_id         = aws_subnet.subnets[each.key].id
#         security_group_id = aws_security_group.sgs[each.key].id
#         gateway           = aws_internet_gateway.igws[each.key]
#         image             = "worker" # Need to create this still
#         master_ip         = locals.ip_map.value[each.value.uuid]
#       }
#   ]
# }

module "tasks"{
  for_each = { for idx, item in jsondecode(file("infra.json")) : idx => item }
  source   = "./tasks"

  uuid              = each.value.uuid
  n                 = each.value.n
  subnet_id         = aws_subnet.subnets[each.key].id
  security_group_id = aws_security_group.sgs[each.key].id
  gateway_id        = aws_internet_gateway.igws[each.key]
}

output "public-ip" {
  value = module.tasks.*
}
