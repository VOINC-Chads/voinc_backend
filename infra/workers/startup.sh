#!/bin/bash
#echo "Instance started at: $timestamp"
sudo yum update -y
sudo yum install -y docker
sudo systemctl start docker
sudo aws ecr get-login-password --region us-east-2 | sudo docker login --username AWS --password-stdin 997625559881.dkr.ecr.us-east-2.amazonaws.com
sudo docker pull 997625559881.dkr.ecr.us-east-2.amazonaws.com/voinc_repo:voinc-worker
TOKEN=$(curl -X PUT "http://169.254.169.254/latest/api/token" -H "X-aws-ec2-metadata-token-ttl-seconds: 21600")
echo "TOKEN: $TOKEN"
INSTANCE_IP=$(curl -H "X-aws-ec2-metadata-token: $TOKEN" -v http://169.254.169.254/latest/meta-data/public-ipv4)
echo "IP: $INSTANCE_IP"
sudo docker run -p 8000:8000 -e IP_ADDR="$INSTANCE_IP" -e ZK_IP_ADDR="${master_ip}" 997625559881.dkr.ecr.us-east-2.amazonaws.com/voinc_repo:voinc-worker