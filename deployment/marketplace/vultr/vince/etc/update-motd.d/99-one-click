#!/bin/sh
#
# Configured as part of the DigitalOcean 1-Click Image build process

myip=$(hostname -I | awk '{print$1}')
cat <<EOF
********************************************************************************

Welcome to Vince server!
To keep this server secure, the UFW firewall is enabled.
All ports are BLOCKED except 22 (SSH), 80 (HTTP), and 443 (HTTPS), 8080 (Vince HTTP)

In a web browser, you can view:
 * The Vince Quickstart guide: https://kutt.it/1click-quickstart

On the server:
  * The default Vince root is located at /var/lib/vince-data
  * Vince is running on ports: 8080 and they are bound to the local interface.

********************************************************************************
  # This image includes version VINCE_VERSION of Vince.
  # See Release notes https://github.com/vinceanalytics/vince/releases/tag/VINCE_VERSION

  # Website:       https://vinceanalytics.com
  # Documentation: https://vinceanalytics.com/guides
  # Vince Github : https://github.com/vinceanalytics/vince

  # Vince config:   /etc/vince/vince.conf

********************************************************************************
EOF
