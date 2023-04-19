resource "aws_instance" "ec2_worker" {

  ami           = "ami-0103f211a154d64a6"
  instance_type = "t2.micro"

  subnet_id                   = var.subnet_id
  vpc_security_group_ids      = [var.security_group_id]

  # Full access to ECR
  iam_instance_profile = "ec2-profile"

  associate_public_ip_address = true

  user_data = templatefile("workers/startup.sh", { master_ip = var.master_ip })

  tags = {
    Name = var.name
  }
}

