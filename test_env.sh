export PORT=80
export LOG_LEVEL=DEBUG

export ENABLE_AUTH=true
export AUTH_USERNAME=esh
export AUTH_PASSWORD=password

export AWS_REGION=us-west-2
export EC2_IMAGE_ID=ami-0698e8523697f84ba # Ubuntu 2204 with CUDA # ami-03f65b8614a860c29 Ubuntu 2204
export EC2_INSTANCE_TYPE=g4dn.xlarge
export EC2_DISK_SIZE=32
export AWS_KEY_NAME=self-host
export LAUNCH_WAIT_TIME=240s
export AWS_SECURITY_GROUP_ID=sg-0456e2fd2eb9ce90b
export AWS_USE_PRIVATE_DNS=true
export EC2_PORT=7860
export EC2_SCRIPT="#!/bin/bash
apt update
apt install -y wget git python3 python3-venv build-essential coreutils
cd /home/ubuntu
# wget https://developer.download.nvidia.com/compute/cuda/12.0.0/local_installers/cuda_12.0.0_525.60.13_linux.run
# sh cuda_12.0.0_525.60.13_linux.run --silent
git clone https://github.com/AUTOMATIC1111/stable-diffusion-webui.git
chown -R ubuntu:ubuntu stable-diffusion-webui/
sysctl -w net.ipv4.ip_unprivileged_port_start=80 >> /etc/sysctl.conf
sudo -u ubuntu nohup bash stable-diffusion-webui/webui.sh --listen"
