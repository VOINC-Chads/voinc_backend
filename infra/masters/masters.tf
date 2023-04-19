resource "aws_instance" "ec2_master" {

  ami           = "ami-0103f211a154d64a6"
  instance_type = "t2.micro"

  subnet_id                   = var.subnet_id
  vpc_security_group_ids      = [var.security_group_id]

  # Full access to ECR
  iam_instance_profile = "ec2-profile"

  associate_public_ip_address = true
  #user_data_base64 = "${base64encode(file("masters/startup.sh"))}"
  user_data = file("masters/startup.sh")


  tags = {
    Name = var.name
  }
}

output "server_public_ip" {
  value = aws_instance.ec2_master.public_ip
}

