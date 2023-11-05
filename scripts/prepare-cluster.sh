#!/bin/bash

#docker
sudo apt-get update
sudo apt-get install ca-certificates curl gnupg
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg

echo \
  "deb [arch="$(dpkg --print-architecture)" signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  "$(. /etc/os-release && echo "$VERSION_CODENAME")" stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt-get update

sudo apt-get install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

#helm
curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3
chmod 700 get_helm.sh
./get_helm.sh

#kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x ./kubectl
sudo mv kubectl /usr/local/bin/

# add user
adduser --disabled-password --gecos "" developer
usermod -aG docker developer && newgrp docker

#minikube
sudo apt install curl wget apt-transport-https -y
curl -Lo minikube https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
chmod +x minikube
mv ./minikube /usr/local/bin/minikube

#su - developer

# run minikube
minikube start --driver=docker --force
minikube addons enable ingress
minikube addons enable ingress-dns
eval $(minikube -p minikube docker-env)

# network
echo "$(minikube ip) minikubeip" >> /etc/hosts
sysctl -w net.ipv4.conf.eth0.route_localnet=1
iptables -t nat -A PREROUTING -d "${EXTERNAL_IP}" -p tcp -m tcp --dport 80 -j DNAT --to-destination 127.0.0.1:80
#kubectl port-forward svc/gateway-svc 80:80 &> /dev/null &  #--address 0.0.0.0

#minikube tunnel -c &> /dev/null &

