#!/bin/bash
################################################
## Prerequisites
chmod +x /root/vultr-helper.sh
. /root/vultr-helper.sh
error_detect_on
install_cloud_init latest

################################################
## Create vince user
groupadd -r vince
useradd -g vince -d /var/lib/vince-data -s /sbin/nologin --system vince

mkdir -p /var/lib/vince-data
chown -R vince:vince /var/lib/vince-data

################################################
## Download vince
wget https://github.com/vinceanalytics/vince/releases/download/v${VINCE_VERSION}/vince_linux-x86_64.tar.gz -O /tmp/vince.tar.gz
tar xvf /tmp/vince.tar.gz -C /usr/bin
chmod +x /usr/bin/vince
chown root:root /usr/bin/vince

################################################
## Install provisioning scripts
mkdir -p /var/lib/cloud/scripts/per-boot/
mkdir -p /var/lib/cloud/scripts/per-instance/

mv /root/setup-per-boot.sh /var/lib/cloud/scripts/per-boot/setup-per-boot.sh
mv /root/setup-per-instance.sh /var/lib/cloud/scripts/per-instance/setup-per-instance.sh

chmod +x /var/lib/cloud/scripts/per-boot/setup-per-boot.sh
chmod +x /var/lib/cloud/scripts/per-instance/setup-per-instance.sh

# Enable Vince on boot
systemctl enable vince.service

################################################
## Prepare server for Marketplace snapshot

clean_system