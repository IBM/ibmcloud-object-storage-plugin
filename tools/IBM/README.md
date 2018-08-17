# How to create K8S secrets
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
