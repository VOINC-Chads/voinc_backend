#!/bin/bash
echo "Instance started at: $timestamp"
sudo yum update -y
sudo yum install -y docker
sudo systemctl start docker
sudo aws ecr get-login-password --region us-east-2 | sudo docker login --username AWS --password-stdin 997625559881.dkr.ecr.us-east-2.amazonaws.com
sudo docker pull 997625559881.dkr.ecr.us-east-2.amazonaws.com/voinc_repo:master
sudo docker stop $(sudo docker ps -q)
sudo docker run -p 8000:8000 997625559881.dkr.ecr.us-east-2.amazonaws.com/voinc_repo:master