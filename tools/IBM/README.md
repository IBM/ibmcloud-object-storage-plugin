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


# How to cleanup s3fs pods stuck in terminating state

Sometimes due to various reasons underlying s3fs process may get killed but the corresponding mount points are not cleaned up. This will cause errors like 
`Transport endpoint is not connected` or the specific pod may remain in `Terminating` state for ever.

In order to clean up those stale s3fs mount points follow the below steps

**Note**: You should have cluster admin access to perform below operations

1. Clone armada-storage-s3fs-plugin repo and move to `tools` directory

   ```
      git clone git@github.ibm.com:alchemy-containers/armada-storage-s3fs-plugin.git  
      cd armada-storage-s3fs-plugin/tools
   ```
2. Fetch cluster config (kubeconfig)

      `$ bx cs cluster-config <cluster-name>`
      
3. Export kubeconfig of your cluster

    `export KUBECONFIG=<armada cluster config file>`
    
4. Create a daemonset to clean the stuck pods if any

    `kubectl apply -f s3fs-mount-cleanup.yaml`
    
5. Verify if daemonset has been succesfully installed

   ```
      root@jupiter-vm360:~/cleanupmounts# kubectl get ds
      NAME                 DESIRED   CURRENT   READY     UP-TO-DATE   AVAILABLE   NODE SELECTOR   AGE
      cleanup-s3fs-mount   3         3         3         3            3           <none>          9s
   ```

6. Once the execution of step 4 is done, wait for 5 minutes and then delete the daemonset using the command

    `kubectl delete -f s3fs-mount-cleanup.yaml`
