# Binary Executable Compilation and Deployment

# Compilation
`build-all.sh` will build both binary executables `s3fs` and `ibmc-s3fs` inside Docker containers that can afterwards be deployed via `kubectl create -f deploy-plugin.yaml`.

# Deployment
The YML files already contain a registry secret. You can create one for your registry via
```
kubectl create secret docker-registry regcred  -n kube-system --docker-server=<REGISTRY_URL> --docker-username=token --docker-password=<REGISTRY_PASSWORD> --docker-email=<EMAIL_ADDRESS>
```
You also want to modify the `image:` fields to include the prefix `<REGISTRY_URL>/<NAMESPACE>/`.

Afterwards, you can deploy the binaries via `kubectl create -f deploy-plugin.yaml` and the plugin via `kubectl create -f deploy-provisioner.yaml`.