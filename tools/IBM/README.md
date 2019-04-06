# How to create K8S secrets?
1. On your local system, log-into `ibmcloud`
   ```
   ibmcloud login -a api.ng.bluemix.net -u nkashyap@in.ibm.com
   ```
2. Export the KUBECONFIG
   ```
   export KUBECONFIG=<armada cluster config file>
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
1. Export the KUBECONFIG
   ```
   export KUBECONFIG=<armada cluster config file>
   ```

2. Execute diagnostic tool
   ```
   $ ./ibm-cos-plugin-diag -h
   usage: ibm-cos-plugin-diag [-h] [--namespace string] [--pvc-name string]
                              [--pod-name string] [--provisioner-log]
                              [--driver-log all|<NODE1>,<NODE2>,...]

   Tool to diagnose object-storage-plugin related issues.

   optional arguments:
     -h, --help            show this help message and exit
     --namespace string, -n string
                           Namespace where the pvc/pod is created. Defaults to
                           "default" namespace
     --pvc-name string     Name of the pvc
     --pod-name string     Name of the pod
     --provisioner-log     Collect provisioner log
     --driver-log all|<NODE1>,<NODE2>,...
                           Name of worker nodes from which driver log needs to be
                           collected seperated by comma(,)
   ```
