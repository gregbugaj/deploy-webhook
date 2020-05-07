#!/usr/bin/env bash

echo 'Installing deploy-webhook'
if  [ $(id -u) = 0 ]; then
   echo "This script must not be run as root, run under 'app-svc' account."
   exit 1
fi


echo 'Stopping existing service'
sudo systemctl stop deploy-webhook.service

echo "Building"

go build
sudo cp deploy-webhook /usr/local/bin
rm ./deploy-webhook

echo "Setup directory '/etc/deploy-webhook/'"
sudo mkdir -p /etc/deploy-webhook
sudo chown app-svc:app-svc /etc/deploy-webhook
cp ./deploy-ansible.sh /etc/deploy-webhook/deploy.sh


#sudo chown app-svc:app-svc /etc/deploy-webhook

echo 'Installing service : deploy-webhook.service'
sudo cp deploy-webhook.service /etc/systemd/system/deploy-webhook.service

echo '------ Service commands --------'
echo 'Service start : sudo systemctl start deploy-webhook.service'
echo 'Service status: sudo systemctl status deploy-webhook.service'

echo 'Starting services'
sudo systemctl start deploy-webhook.service
sudo systemctl status deploy-webhook.service
