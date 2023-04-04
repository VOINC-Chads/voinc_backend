locals {
  my_list_of_dicts = jsondecode(var.images)
}

resource "aws_instance" "my_instances" {
  for_each = {
    for idx, dict in local.my_list_of_dicts :
      idx => dict
  }

  ami           = data.aws_ami.amazon_linux.id
  instance_type = "t2.micro"

  # Full access to ECR
  vpc_security_group_ids = ["${aws_security_group.default-sg.id}"]
  iam_instance_profile = "ec2-profile"

  # All on same subnet, each with different private IP
  subnet_id              = aws_subnet.voinc-subnet.id
  private_ip             = "10.0.1.${each.key}"
  
  user_data = <<EOF
         #!/bin/bash
         docker run -d --name my-container-${each.key} ${each.value["image"]}  > /dev/null


         # Wait for the Docker container to finish running the program
         # while docker ps --format '{{.Names}}' | grep -q 'container-${each.key}'; do
         #    sleep 1
         #done

         # Send the final result back to a client using an HTTP response
         #INSTANCE_ID=\$(curl -s http://169.254.169.254/latest/meta-data/instance-id)
         #FINAL_RESULT=\$(docker logs container-${each.key} | tail -1)
         # NEED WHERE TO SEND BACK
         #curl -X POST -d "{\\"result\\": \\"\$FINAL_RESULT\\"}" http://client-server.com/result --header "Content-Type:application/json"

         # Terminate the EC2 instance
         #aws ec2 terminate-instances --instance-ids \$INSTANCE_ID --region ${var.region}
         EOF

  tags = {
    Name = "MyInstance-${each.key}"
  }
}
