#!/bin/bash

# Install uid map
sudo apt update
sudo apt install -y uidmap

# Install Docker
curl -fsSL https://get.docker.com | sudo sh
curl -fsSL https://get.docker.com/rootless | sh

# Setup port for non-root users
sudo sysctl -w net.ipv4.ip_unprivileged_port_start=80 >> /etc/sysctl.conf