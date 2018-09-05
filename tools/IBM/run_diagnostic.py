#! /usr/bin/python
# vim: tabstop=4 shiftwidth=4 softtabstop=4
# ******************************************************************************
# * Licensed Materials - Property of IBM
# * IBM Cloud Container Service, 5737-D43
# * (C) Copyright IBM Corp. 2017, 2018 All Rights Reserved.
# * US Government Users Restricted Rights - Use, duplication or
# * disclosure restricted by GSA ADP Schedule Contract with IBM Corp.
# ******************************************************************************

import argparse
import datetime
import json
import math
import os
import subprocess
import sys
import threading
import time
from Queue import Queue
from threading import Thread

node_list = [ ]
nodes_available = ""
nodeqone = Queue(maxsize=0)
nodeqtwo = Queue(maxsize=0)

### Create daemonset yaml file in local machine
##
#
yamlfile = '''\
apiVersion: extensions/v1beta1
kind: DaemonSet
metadata:
  name: s3fs-diagnostic
spec:
  template:
    metadata:
      labels:
        name: s3fs-diagnostic
    spec:
      hostNetwork: true
      containers:
      - image: ambikanair/s3fs-diagnostic-tool:3
        securityContext:
          privileged: true
        name: s3fs-diagnostic
        env:
        - name: HOST_IP
          valueFrom:
           fieldRef:
             fieldPath: status.hostIP
        volumeMounts:
        - mountPath: /host
          name: root-fs
        - mountPath: /run/systemd
          name: host-systemd
        - mountPath: "/var/log/"
          name: s3fs-log
      volumes:
      - name: root-fs
        hostPath:
          # directory location on host
          path: /
      - name: host-systemd
        hostPath:
          path: /run/systemd
      - name: s3fs-log
        hostPath:
          path: /var/log/
'''
#
##
###

with open('diagnostic_daemon.yaml', 'w') as the_file:
    the_file.write(yamlfile)

    class cmdHandler:

        # Internal method to execute a command
        # RC 0 - Cmd execution with zero return code
        # RC 1 - Cmd executed with non zero return code
        # RC 2 - Runtime error / exception
        def cmd_run(self, cmd):
            cmd_output = ''
            cmd_err = ''
            rc = 0

            #print ("CommmandExec \"{0}\"".format(cmd))
            try:
                process = subprocess.Popen(cmd, stdout=subprocess.PIPE,
                                       stderr=subprocess.PIPE, shell=True)
                process.wait()
                (cmd_output, cmd_err) = process.communicate()
            except Exception as err:
                print ("Command \"{0}\" execution failed. ERROR: {1}".format(cmd, str(err)))
                cmd_err = "Command execution failed"
                rc = 2
            else:
                if process.returncode == 0:
                    rc = 0
                else:
                    rc = 1
                if cmd_err:
                   print ("{0}\nERROR: {1}".format(cmd, cmd_err.strip()))
            return (rc, cmd_output, cmd_err)

cmdHandle = cmdHandler()

def executeBasicChecks():

    print "\n****ibmcloud-object-storage-plugin pod status****"
    cmd = "kubectl get pods -n kube-system -o wide| grep object-storage-plugin| awk '{print $3}'"
    (rc, cmd_out, cmd_err) = cmdHandle.cmd_run(cmd)
    if not cmd_err.strip():
        if not cmd_out.strip():
            print "ERROR: ibmcloud-object-storage-plugin pod not found"
        else:
            print "ibmcloud-object-storage-plugin pod is in \"{0}\" state.".format(cmd_out.strip())

    print "\n****Checking storage class details****"
    cmd = "kubectl get sc | grep s3fs"
    (rc, cmd_out, cmd_err) = cmdHandle.cmd_run(cmd)
    if not cmd_err.strip():
        print cmd_out.strip()

    print "\n****Total number of storage classes created****"
    cmd = "kubectl get sc | grep s3fs | wc -l"
    (rc, cmd_out, cmd_err) = cmdHandle.cmd_run(cmd)
    if not cmd_err.strip():
        print cmd_out.strip()

    print "\n****Checking iam-endpoint****"
    cmd = "kubectl get sc ibmc-s3fs-standard-cross-region -o jsonpath='{.parameters.ibm\.io/iam-endpoint}' "
    (rc, cmd_out, cmd_err) = cmdHandle.cmd_run(cmd)
    if not cmd_err.strip():
        if not cmd_out.strip():
            print "ERROR: iam-endpoint is empty"
        else:
            print "iam-endpoint is set"

    print "\n****Checking object-storage-endpoint****"
    cmd = "kubectl get sc ibmc-s3fs-standard-cross-region -o jsonpath='{.parameters.ibm\.io/object-store-endpoint}'"
    (rc, cmd_out, cmd_err) = cmdHandle.cmd_run(cmd)
    if not cmd_err.strip():
        if ("NA" == cmd_out.strip()) or not cmd_out.strip():
            print "ERROR: object-storage-endpoint is empty"
        else:
            print "object-storage-endpoint is set"

    print "\n****Checking object-store-storage-class****"
    cmd = "kubectl get sc ibmc-s3fs-standard-cross-region -o jsonpath='{.parameters.ibm\.io/object-store-storage-class}'"
    (rc, cmd_out, cmd_err) = cmdHandle.cmd_run(cmd)
    if not cmd_err.strip():
        if ("NA" == cmd_out.strip()) or not cmd_out.strip():
            print "ERROR: object-store-storage-class is empty."
        else:
            print "object-store-storage-class is set"

    print "\nChecking ServiceAccounts, ClusterRoles and ClusterRoleBindings are created"
    print "****ServiceAccounts****"
    cmd = "kubectl get sa -n kube-system | grep ibmcloud-object-storage-driver | awk '{print $1}'"
    (rc, cmd_out, cmd_err) = cmdHandle.cmd_run(cmd)
    if not cmd_err.strip():
        if "ibmcloud-object-storage-driver" == cmd_out.strip():
            print cmd_out.strip() + " is created"
        else:
            print "ERROR: ibmcloud-object-storage-driver is not created"

    cmd = "kubectl get sa -n kube-system | grep ibmcloud-object-storage-plugin | awk '{print $1}'"
    (rc, cmd_out, cmd_err) = cmdHandle.cmd_run(cmd)
    if not cmd_err.strip():
        if "ibmcloud-object-storage-plugin" == cmd_out.strip():
            print cmd_out.strip() + " is created"
        else:
            print "ERROR: ibmcloud-object-storage-plugin is not created"

    print "\n****ClusterRole****"
    cmd = "kubectl get ClusterRole | grep ibmcloud-object-storage-plugin | awk '{print $1}'"
    (rc, cmd_out, cmd_err) = cmdHandle.cmd_run(cmd)
    if not cmd_err.strip():
        if "ibmcloud-object-storage-plugin" == cmd_out.strip():
            print cmd_out.strip() + " is created"
        else:
            print "ERROR: ibmcloud-object-storage-plugin is not created"

    cmd = "kubectl get ClusterRole | grep ibmcloud-object-storage-secret-reader | awk '{print $1}'"
    (rc, cmd_out, cmd_err) = cmdHandle.cmd_run(cmd)
    if not cmd_err.strip():
        if "ibmcloud-object-storage-secret-reader" == cmd_out.strip():
            print cmd_out.strip() + " is created"
        else:
            print "ERROR: ibmcloud-object-storage-secret-reader is not created"

    print "\n****ClusterRoleBinding****"
    cmd = "kubectl get ClusterRoleBinding | grep ibmcloud-object-storage-plugin | awk '{print $1}'"
    (rc, cmd_out, cmd_err) = cmdHandle.cmd_run(cmd)
    if not cmd_err.strip():
        if "ibmcloud-object-storage-plugin" == cmd_out.strip():
            print cmd_out.strip() + " is created"
        else:
            print "ERROR: ibmcloud-object-storage-plugin is not created"

    cmd = "kubectl get ClusterRoleBinding | grep ibmcloud-object-storage-secret-reader | awk '{print $1}'"
    (rc, cmd_out, cmd_err) = cmdHandle.cmd_run(cmd)
    if not cmd_err.strip():
        if "ibmcloud-object-storage-secret-reader" == cmd_out.strip():
            print cmd_out.strip() + " is created"
        else:
            print "ERROR: ibmcloud-object-storage-secret-reader is not created"

def getNodeList(q):
    #cmd = "kubectl get nodes -o jsonpath='{.items[*].status.addresses[?(@.type==\"Hostname\")].address}'"
    cmd = "kubectl get nodes | grep -w \"Ready\" | awk '{print $1}'"
    (rc, cmd_out, cmd_err) = cmdHandle.cmd_run(cmd)
    if not cmd_err.strip():
        global node_list
        node_list = cmd_out.strip().split()
        for x in node_list:
            q.put(x)

def backupDriverLog(q, lockone):
     while True:
         #print threading.current_thread()
         lockone.acquire()
         x = q.get()
         print "Copying driver log from node {0}".format(x)
         cmd = "kubectl get pods -o wide | grep s3fs-diagnostic | grep  \"{0}\" |awk '{{print $1}}'".format(x)
         (rc, cmd_out, cmd_err) = cmdHandle.cmd_run(cmd)
         if not cmd_err.strip():
             podname =  cmd_out.strip()
             cmd = "kubectl cp {0}:var/log/ibmc-s3fs.log {1}-ibmc-s3fs.log".format(podname, x)
             (rc, cmd_out, cmd_err) = cmdHandle.cmd_run(cmd)

         q.task_done()
         lockone.release()

def backupdiagnosticLogs(q, locktwo):
     while True:
         #print threading.current_thread()
         locktwo.acquire()
         x = q.get()
         print "Copying diagnostic log from node {0}".format(x)
         #print(str(datetime.datetime.now()))
         cmd = "kubectl get pods -o wide | grep s3fs-diagnostic | grep  \"{0}\" |awk '{{print $1}}'".format(x)
         (rc, cmd_out, cmd_err) = cmdHandle.cmd_run(cmd)
         if not cmd_err.strip():
             podname =  cmd_out.strip()
             cmd = "kubectl cp {0}:var/log/checkMountStatus.log {1}-s3fsMountStatus.log".format(podname, x)
             (rc, cmd_out, cmd_err) = cmdHandle.cmd_run(cmd)
             if not cmd_err.strip():
                 cmd ="cat {0}-s3fsMountStatus.log | grep -e ERROR".format(x)
                 (rc, cmd_out, cmd_err) = cmdHandle.cmd_run(cmd)
                 if (not cmd_err.strip()) and cmd_out.strip():
                     print cmd_out.strip()

         q.task_done()
         locktwo.release()

def check_daemonset_state(ds_name):
    attempts = 0
    global nodes_available

    cmd = "kubectl get nodes | grep -w \"Ready\" | wc -l"
    (rc, cmd_out, cmd_err) = cmdHandle.cmd_run(cmd)
    nodes_available = cmd_out.strip()

    while True:
        attempts = attempts + 1
        cmd = "kubectl get ds | grep \"{0}\" | awk '{{print $4}}'".format(ds_name)
        (rc, cmd_out, cmd_err) = cmdHandle.cmd_run(cmd)
        ds_status_ready = cmd_out.strip()

        if ds_status_ready == nodes_available:
           print ds_name + " is  running and available ds instances:" +  ds_status_ready
           break

        if attempts > 30 :
            print ds_name + "were not running well. Instances Desired:" + nodes_available + ", Instances Available:" + ds_status_ready
            cmd = "kubectl get ds {0}".format(ds_name)
            (rc, cmd_out, cmd_err) = cmdHandle.cmd_run(cmd)
            if not cmd_err.strip():
                print cmd_out.strip()
            sys.exit(1)

        print "DS:{0}, Desired:{1}, Available:{2} Sleeping 10 seconds".format(ds_name, nodes_available, ds_status_ready)
        time.sleep(10)

def cleanup_daemonset(ds_name):
    print "****Deleting daemonset {0}****".format(ds_name)

    cmd = "kubectl delete ds {0}".format(ds_name)
    (rc, cmd_out, cmd_err) = cmdHandle.cmd_run(cmd)
    if not cmd_err.strip():
        print cmd_out.strip()
    else:
        sys.exit(1)


def main():
    parser = argparse.ArgumentParser(description='Tool to diagnose object-storage-plugin related issues.')
    parser.add_argument("--namespace", "-n", default="default", dest="ns", metavar="string",
                        help="Namespace where the pvc/pod is created. Defaults to \"default\" namespace")
    parser.add_argument("--pvc-name", dest="pvc", metavar="string", help="Name of the pvc")
    parser.add_argument("--pod-name", dest="pod", metavar="string", help="Name of the pod")
    parser.add_argument("--provisioner-log", dest="provisioner_log", action="store_true",
                        help="Collect provisioner log")
    parser.add_argument("--driver-log", dest="driver_log", metavar="all|<NODE1>,<NODE2>,...",
                        action='append',
                        help="Name of worker nodes from which driver log needs to be collected seperated by comma(,)")

    args = parser.parse_args()
    if "KUBECONFIG" not in os.environ:
        print "export KUBECONFIG before running this tool.\nExiting!!!"
        sys.exit(1)
    print "****Checking whether cluster nodes are accessible****"
    cmd =  "kubectl get nodes"
    (rc, cmd_out, cmd_err) = cmdHandle.cmd_run(cmd)
    if cmd_err.strip():
        print "Cluster nodes are not accessible. Please check KUBECONFIG env variable."
        sys.exit(1)
    else:
        print "Success!!!"

    executeBasicChecks()

    if args.pvc:
        print "\n****Checking input PVC status****"
        cmd =  "kubectl describe pvc -n " + args.ns + " " + args.pvc
        (rc, cmd_out, cmd_err) = cmdHandle.cmd_run(cmd)
        if not cmd_err.strip():
            print cmd + "\n" + cmd_out

    if args.pod:
        print "\n****Checking input POD status****"
        cmd =  "kubectl describe pod -n " + args.ns + " " + args.pod
        (rc, cmd_out, cmd_err) = cmdHandle.cmd_run(cmd)
        if not cmd_err.strip():
            print cmd + "\n" + cmd_out

    # Collect provisioner log
    if args.pvc or args.provisioner_log:
        print "\n****Collecting provisioner log****"
        cmd = "kubectl get pods -n kube-system -o wide| grep object-storage-plugin| awk '{print $1,$3}'"
        (rc, cmd_out, cmd_err) = cmdHandle.cmd_run(cmd)
        if not cmd_err.strip():
            if not cmd_out.strip():
                print "ERROR: ibmcloud-object-storage-plugin pod not found. Provisioner log cannot be fetched."
            else:
                plugin_pod = cmd_out.strip().split(" ")
                plugin_pod_name = plugin_pod[0]
                plugin_pod_status = plugin_pod[1]
                if plugin_pod_status == "Running":
                    cmd = "kubectl logs {0} -n kube-system > s3provisioner.log".format(plugin_pod_name)
                    (rc, cmd_out, cmd_err) = cmdHandle.cmd_run(cmd)
                    if not cmd_err.strip():
                        print "Success!!!"
                else:
                    print "ERROR: ibmcloud-object-storage-plugin is in \"{0}\" state. Provisioner log cannot be fetched.".format(plugin_pod_status)
        else:
            print "Provisioner log cannot be fetched."

    cmd = "kubectl apply -f diagnostic_daemon.yaml"
    (rc, cmd_out, cmd_err) = cmdHandle.cmd_run(cmd)
    if cmd_err.strip():
        print "Exiting!!!"
        sys.exit(1)

    # wait for the daemonset pod to reach running state
    print "\n*****Check s3fs-diagnostic daemonset status*****"
    check_daemonset_state("s3fs-diagnostic")

    #print "\nExecuting checks inside each worker node"
    global nodeqone
    global nodeqtwo
    global node_list

    flag = 0
    # Collect driver logs based on --driver-log option
    if args.driver_log:
        for x in args.driver_log:
            if "all" in x:
                flag = 1
                getNodeList(nodeqone)
                break
            else:
                if "," in x:
                    input_workers = x.split(',')
                    for ip in input_workers:
                        node_list.append(ip.strip())
                else:
                    node_list.append(x.strip())

    if args.pod and (flag == 0):
        cmd =  "kubectl get pods -o jsonpath='{.status.hostIP}' -n " + args.ns + " " + args.pod
        (rc, cmd_out, cmd_err) = cmdHandle.cmd_run(cmd)
        if (not cmd_err.strip()) and (cmd_out.strip() not in node_list):
            node_list.append(cmd_out.strip())

    if (flag == 0) and (len(node_list) != 0):
        for x in node_list:
            nodeqone.put(x)

    # creating a lock
    lockone = threading.Lock()
    locktwo = threading.Lock()

    # Set thread count equals to 1/3rd of worker nodes
    NUMBER_OF_WORKERS = int(math.ceil(int(nodes_available)/float(3)))

    for i in range(NUMBER_OF_WORKERS):
        worker = Thread(target=backupDriverLog, args=(nodeqone,lockone, ))
        worker.setDaemon(True)
        worker.start()

    nodeqone.join()
    time.sleep(10)

    getNodeList(nodeqtwo)
    for j in range(NUMBER_OF_WORKERS):
        workertwo = Thread(target=backupdiagnosticLogs, args=(nodeqtwo,locktwo, ))
        workertwo.setDaemon(True)
        workertwo.start()
    nodeqtwo.join()

    time.sleep(10)

    cmd = "uname"
    (rc, cmd_out, cmd_err) = cmdHandle.cmd_run(cmd)
    os_name = cmd_out.strip()

    if os_name == "Darwin":
        compress_cmd = "zip -r -X s3fs_diagnostic_logs.zip "
        file_extension = ".zip"
    else:
        compress_cmd = "tar -cvf s3fs_diagnostic_logs.tar "
        file_extension = ".tar"

    if (args.pvc or args.provisioner_log) and (args.driver_log or args.pod):
        cmd = compress_cmd + "s3provisioner.log *ibmc-s3fs.log *s3fsMountStatus.log"
        (rc, cmd_out, cmd_err) = cmdHandle.cmd_run(cmd)
        if not cmd_err.strip():
            cmd = "pwd"
            (rc, cmd_out, cmd_err) = cmdHandle.cmd_run(cmd)
            print "Successfully created log archive \"s3fs_diagnostic_logs" + file_extension \
            + "\" and saved it to: " + cmd_out.strip() + "/s3fs_diagnostic_logs" + file_extension

    elif args.driver_log or args.pod:
        cmd = compress_cmd + "*ibmc-s3fs.log *s3fsMountStatus.log"
        (rc, cmd_out, cmd_err) = cmdHandle.cmd_run(cmd)
        if not cmd_err.strip():
            cmd = "pwd"
            (rc, cmd_out, cmd_err) = cmdHandle.cmd_run(cmd)
            print "Successfully created log archive \"s3fs_diagnostic_logs" + file_extension \
            + "\" and saved it to: " + cmd_out.strip() + "/s3fs_diagnostic_logs" + file_extension

    elif args.pvc or args.provisioner_log:
        cmd = compress_cmd + "s3provisioner.log *s3fsMountStatus.log"
        (rc, cmd_out, cmd_err) = cmdHandle.cmd_run(cmd)
        if not cmd_err.strip():
            cmd = "pwd"
            (rc, cmd_out, cmd_err) = cmdHandle.cmd_run(cmd)
            print "Successfully created log archive \"s3fs_diagnostic_logs" + file_extension \
            + "\" and saved it to: " + cmd_out.strip() + "/s3fs_diagnostic_logs" + file_extension

    else:
        cmd = compress_cmd + "*s3fsMountStatus.log"
        (rc, cmd_out, cmd_err) = cmdHandle.cmd_run(cmd)
        if not cmd_err.strip():
            cmd = "pwd"
            (rc, cmd_out, cmd_err) = cmdHandle.cmd_run(cmd)
            print "Successfully created log archive \"s3fs_diagnostic_logs" + file_extension \
            + "\" and saved it to: " + cmd_out.strip() + "/s3fs_diagnostic_logs" + file_extension

    #cmd = "ls | grep -w \".*s3.*.log\" | xargs -d\"\\n\" rm -rf"
    cmd = "ls | grep -w \".*s3.*.log\" | xargs -L1 rm -rf"
    (rc, cmd_out, cmd_err) = cmdHandle.cmd_run(cmd)

    # Delete s3fs-diagnostic daemonset
    cleanup_daemonset("s3fs-diagnostic")

if __name__ == "__main__":
    main()
