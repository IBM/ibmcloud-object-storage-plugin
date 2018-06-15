#!/bin/sh

apt-get install -q -y libcurl4-openssl-dev
systemctl restart kubelet.service
