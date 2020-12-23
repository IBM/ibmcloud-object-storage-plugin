# How to create K8S secrets?
1. On your local system, log-into `ibmcloud`
   ```
   ibmcloud login -a cloud.ibm.com -u nkashyap@in.ibm.com
   ```
2. Set the cluster KUBECONFIG
   ```
   ibmcloud ks cluster config -c <cluster_name/cluster_id>
   ```
3. Create K8S secrets
   ```
   $ ./create-k8s-secret -h
   Usage: ./create-k8s-secret <auth-type> <service-key> <secret-name> <namespace>
      Where:-
      - <auth-type>   : iam or hmac
      - <service-key> : service key, to get the list of keys execute
                        ibmcloud resource service-keys --instance-name <instance name>
      - <secret-name> : secret name to be assigned
      - <namespace>   : K8S namespace under which the secret to be created
   ```

# How to execute diagnostic tool?
1. Set the cluster KUBECONFIG
   ```
   ibmcloud ks cluster config -c <cluster_name/cluster_id>
   ```
2. Clone the repo   
`git clone https://github.com/IBM/ibmcloud-object-storage-plugin.git`
 
    and navigate to diagnostic tool   
 `cd ibmcloud-object-storage-plugin/tools/IBM`

3. Execute diagnostic tool
   ```
   $ ./ibm-cos-plugin-diag -h
   usage: ibm-cos-plugin-diag [-h] [--namespace string] [--pvc-name string]
                              [--pod-name string] [--provisioner-log]
                              [--driver-log --node all|<NODE1>,<NODE2>,...]

   Tool to diagnose object-storage-plugin related issues.

   optional arguments:
     -h, --help            show the help message and exit
     --namespace string, -n string
                           Namespace where the pvc/pod is created. Defaults to
                           "default" namespace
     --pvc-name string     Name of the pvc which need to be debugged
     --pod-name string     Name of the pod which need to be debugged
     --provisioner-log     Collect provisioner logs
     --driver-log --node all|<NODE1>,<NODE2>,...
                           Name of worker nodes from which driver logs needs to be
                           collected seperated by comma(,)
   ```
   
4. Example commands to run diagnostic tool

    To run the sanity check for cos s3fs plugin install
    `./ibm-cos-plugin-diag`
    
    To inspect a pvc   
    `./ibm-cos-plugin-diag --namespace default --pvc-name <pvc_name>`
    
    To fetch provisioner logs  
    `./ibm-cos-plugin-diag --namespace default  --provisioner-log`
    
    To fetch the driver logs for all nodes  
    `./ibm-cos-plugin-diag --driver-log --node all`
    
    To fetch the driver logs for a specific node  
    `./ibm-cos-plugin-diag --driver-log --node <node_name1>,<node_name1>`

5. The log archive will be created under the same directory from where the script is run.   
`ibm-cos-diagnostic-logs.zip`


### Sample output from diagnostic tool

```
Bhagyashrees-MacBook-Pro:IBM bhagyashree$ ./ibm-cos-plugin-diag --provisioner-log --driver-log --node all
creating the diagnostic_daemon.yaml
****Checking access to cluster****
> nodes are accessible... ok

****ibmcloud-object-storage-plugin pod status****
> ibmcloud-object-storage-plugin pod is in "Running" state... ok

****ibmcloud-object-storage-driver pods status****
> All ibmcloud-object-storage-driver pods are in "Running" state... ok

****Default storage class list****
> ibmc-s3fs-cold-cross-region            ibm.io/ibmc-s3fs     Delete          Immediate           false                  6d2h
> ibmc-s3fs-cold-regional                ibm.io/ibmc-s3fs     Delete          Immediate           false                  6d2h
> ibmc-s3fs-flex-cross-region            ibm.io/ibmc-s3fs     Delete          Immediate           false                  6d2h
> ibmc-s3fs-flex-perf-cross-region       ibm.io/ibmc-s3fs     Delete          Immediate           false                  6d2h
> ibmc-s3fs-flex-perf-regional           ibm.io/ibmc-s3fs     Delete          Immediate           false                  6d2h
> ibmc-s3fs-flex-regional                ibm.io/ibmc-s3fs     Delete          Immediate           false                  6d2h
> ibmc-s3fs-standard-cross-region        ibm.io/ibmc-s3fs     Delete          Immediate           false                  6d2h
> ibmc-s3fs-standard-perf-cross-region   ibm.io/ibmc-s3fs     Delete          Immediate           false                  6d2h
> ibmc-s3fs-standard-perf-regional       ibm.io/ibmc-s3fs     Delete          Immediate           false                  6d2h
> ibmc-s3fs-standard-regional            ibm.io/ibmc-s3fs     Delete          Immediate           false                  6d2h
> ibmc-s3fs-vault-cross-region           ibm.io/ibmc-s3fs     Delete          Immediate           false                  6d2h
> ibmc-s3fs-vault-regional               ibm.io/ibmc-s3fs     Delete          Immediate           false                  6d2h

****Inspecting Storage-class "ibmc-s3fs-standard-cross-region"****
> iam-endpoint is set to "https://iam.bluemix.net"
> object-store-endpoint is set to "https://s3.private.dal.us.cloud-object-storage.appdomain.cloud"
> object-store-storage-class is set to "us-standard"
> provisioner is set to "ibm.io/ibmc-s3fs"

****ServiceAccounts****
> ibmcloud-object-storage-driver is defined... ok
> ibmcloud-object-storage-plugin is defined... ok

****ClusterRole****
> ibmcloud-object-storage-plugin is defined... ok
> ibmcloud-object-storage-secret-reader is defined... ok

****ClusterRoleBinding****
> ibmcloud-object-storage-plugin is defined... ok
> ibmcloud-object-storage-secret-reader is defined... ok

creating the deamonset to fetch driver logs
> DS: ibm-cos-plugin-diag, Desired: 2, Available: 0 Sleeping 10 seconds
> 2 instances of ibm-cos-plugin-diag is running

*****Collecting ibm-cos-plugin-diag logs*****
> th-1: Copying diagnostic log from node 10.185.208.242
> th-0: Copying diagnostic log from node 10.185.208.213

*****Analyzing ibm-cos-plugin-diag logs*****
> No Error

****Collecting provisioner log****
> Collected provisioner log... ok

*****Collecting driver logs*****
> th-0 Copying driver log from node 10.185.208.213
 > th-1 Copying driver log from node 10.185.208.242

Successfully created log archive "ibm-cos-diagnostic-logs.zip"
and saved as: /Users/bhagyashree/Documents/ObjectStorage/diagnosticTool/bha-diag/ibmcloud-object-storage-plugin/tools/IBM/ibm-cos-diagnostic-logs.zip

****Deleting daemonset default/ibm-cos-plugin-diag****
daemonset.apps "ibm-cos-plugin-diag" deleted

Bhagyashrees-MacBook-Pro:IBM bhagyashree$ ls
README.md			diag-util			ibm-cos-plugin-diag
create-k8s-secret		ibm-cos-diagnostic-logs.zip	run_diagnostic.py
```
