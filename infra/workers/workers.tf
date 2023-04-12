resource "aws_instance" "ec2_worker" {

  ami           = "ami-0103f211a154d64a6"
  instance_type = "t2.micro"

  subnet_id                   = var.subnet_id
  vpc_security_group_ids      = [var.security_group_id]

  # Full access to ECR
  iam_instance_profile = "ec2-profile"

  associate_public_ip_address = false
  
  user_data = <<EOF
         #!/bin/bash
         sudo yum update -y
         sudo yum install docker
         sudo systemctl start docker
         sudo aws ecr get-login-password --region us-east-2 | sudo docker login --username AWS --password-stdin 997625559881.dkr.ecr.us-east-2.amazonaws.com
         sudo docker pull 997625559881.dkr.ecr.us-east-2.amazonaws.com/voinc_repo:${var.image}
         docker run -p 8000:8000 ${var.image} --ip ${var.master_ip}
         EOF

  tags = {
    Name = var.name
  }
}

