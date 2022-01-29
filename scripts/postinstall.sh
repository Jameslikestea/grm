#!/bin/sh

ssh-keygen -b 2048 -f /etc/grmpkg_hostkey -q -N=""
chown grmpkg:grmpkg /etc/grmpkg_hostkey*
touch /var/log/grmpkg.log
chown grmpkg:grmpkg /var/log/grmpkg.log
chmod 644 /var/log/grmpkg.log