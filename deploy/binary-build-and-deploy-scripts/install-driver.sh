#!/bin/bash

set -ex
DRIVER_LOCATION="/host/usr/libexec/kubernetes/kubelet-plugins/volume/exec/ibm~ibmc-s3fs"
KUBELET_SVC_CONFIG="/host/lib/systemd/system/kubelet.service"

cp /root/bin/s3fs /host/usr/local/bin/
cp /root/bin/install-dep.sh /host/root/
chmod +x /host/local/bin/s3fs /host/root/install-dep.sh 

if [ -e "$DRIVER_LOCATION/ibmc-s3fs" ]
then
	mv /root/bin/ibmc-s3fs $DRIVER_LOCATION/
	chmod +x $DRIVER_LOCATION/ibmc-s3fs
else
	mkdir -p $DRIVER_LOCATION
        cp /root/bin/ibmc-s3fs $DRIVER_LOCATION/
	chmod +x $DRIVER_LOCATION/ibmc-s3fs

	# disable enable-controller-attach-detach
	grep enable-controller-attach-detach $KUBELET_SVC_CONFIG || \
	sed -i '/--api-servers=/a  \\t --enable-controller-attach-detach=false \\'  $KUBELET_SVC_CONFIG
fi

ssh-keygen -N "" -f /root/.ssh/id_rsa

mkdir -p /host/root/.ssh/
if [ -f /host/root/.ssh/authorized_keys ]; then
   cp -p /host/root/.ssh/authorized_keys /host/root/.ssh/authorized_keys_original
fi
cat /root/.ssh/id_rsa.pub >> /host/root/.ssh/authorized_keys
chmod 700 /host/root/.ssh/
chmod 600 /host/root/.ssh/authorized_keys

touch /host/etc/ssh/sshd_config
sed -i 's/PermitRootLogin no/PermitRootLogin yes/g' /host/etc/ssh/sshd_config
/root/bin/systemutil -service ssh.service

ssh -o StrictHostKeyChecking=no root@localhost bash /root/install-dep.sh

sed -i 's/PermitRootLogin yes/PermitRootLogin no/g' /host/etc/ssh/sshd_config
/root/bin/systemutil -service ssh.service

if [ -f /host/root/.ssh/authorized_keys_original ]; then
   cp -p /host/root/.ssh/authorized_keys_original /host/root/.ssh/authorized_keys
   rm -f /host/root/.ssh/authorized_keys_original
fi

set +ex

tail -f /dev/null
