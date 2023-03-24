# Deploying the service on a locl kind cluster

Except for the prerequisites, these steps are repeatable in the Makefile in the root of the repository.

## Step 0: Prerequisites
Install kind, e.g. via [homebrew](https://formulae.brew.sh/formula/kind).

Install [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl-macos/)

## Step 1: kind setup

> From root, you can run `make kind-create` to create the cluster and apply the ingress controller.

run `kind create cluster --name=authz --config=kind.yml`

This applies the kind.yml from this folder.

Notable things in `kind.yml`:

* `extraPortMappings`: Port-forwarding from host to ingress controller. Allows to make requests to the cluster from localhost
* `node-labels`: only allow the ingress controller to run on a specific node matching the label selector

## Step 2: Use the nginx ingress controller

> From root, you can run `make kind-create` to create the cluster and apply the ingress controller.

The current nginx ingress controller manifest for kind can be found [here at the ingress-nginx repo](https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml).

It contains kind-specific patches to forward the host ports to the ingress controller.

Apply it to the cluster via 

`kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml`

Then wait until the nginx controller is running

`kubectl get pods --namespace ingress-nginx` should show an ingress-nginx-controller with status `Running` like this:

```bash
NAME                                       READY   STATUS      RESTARTS   AGE
ingress-nginx-admission-create-pc9d7       0/1     Completed   0          38s
ingress-nginx-admission-patch-bmr4s        0/1     Completed   0          38s
ingress-nginx-controller-58c49c4db-7g8qf   1/1     Running     0          38s
```

## Step 3: Deploy the app

> From root, you can run `make kind-deploy` to do this.

run 

`kubectl apply -f authz.yml`

authz.yml defines:
* a pod with a single container and its image 
* a service that acts as an intermediate between the pod and the ingress controller. The service exposes the pod’s default port (8080 for now).
* an entrypoint/ingress that sits in front of multiple services in the cluster to send the localhost request to the pod.

## Step 4: check if pods and services are running

To check if the pods and services are running, use 

`kubectl get pods,svc`

the result should look like this:

```bash
NAME            READY   STATUS    RESTARTS   AGE
pod/authz-app   1/1     Running   0          24s

NAME                    TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
service/authz-service   ClusterIP   10.96.29.111   <none>        8080/TCP   70s
service/kubernetes      ClusterIP   10.96.0.1      <none>        443/TCP    9m6s

```
## Step 5: Call the service

Navigate to http://localhost:8080/whoever%20is%20reading%20this or run a curl to the http://localhost:8080/ endpoint to see the service runs and returns a greeting as expected. 

## Step 6: Remove everything

Shut down the cluster using

`kind delete cluster --name=authz`
 

