# Istio-Linkerd Service Mesh Demo
- [Istio-Linkerd Service Mesh Demo](#istio-linkerd-service-mesh-demo)
    - [About](#about)
    - [Requirements](#requirements)
    - [Services](#services)
    - [Demo](#demo)
        - [Setup](#setup)
        - [Exercise 1 - Basic Service Mesh](#exercise-1---basic-service-mesh)
            - [Istio](#istio)
            - [Linkerd](#linkerd)
        - [Exercise 2 - Canary Deployment](#exercise-2---canary-deployment)
            - [Istio](#istio)
            - [Linkerd](#linkerd)
        - [Exercise 3 - Error Handling](#exercise-3---error-handling)
            - [Istio](#istio)
            - [Linkerd](#linkerd)
        - [Exercise 4 - Stealth Deployment](#exercise-4---stealth-deployment)
            - [Istio](#istio)
            - [Linkerd](#linkerd)
## About

This repo contains a demo showing the usage of both linkerd and istio in kubernets. It contains a few microservices that will run in the service mesh.
## Requirements

* minikube running kubernetes >= v1.8.0
  * Ingress installed on minikube `minikube addons enable ingress`
* kubectl
* helm
* helmfile
* siege
* Docker >= v17.05.0-ce
* namerctl
  * `go get -u github.com/linkerd/namerctl`
  * `go install github.com/linkerd/namerctl`

## Services
There are three services:
1. A words service which generates random words
1. A simon says service which calls the words service
1. A capitalization service which calls either the simon service or the words service

## Demo
### Setup
1. First you'll want a running k8s cluster on minikube. This is as simple as running the command `minikube start`
2. One the cluster is up and running, you'll need a tiller pod running for helm to talk to. Run `helm init` to kick this off
3. Install the Istio components with the helmfile provided in the `istio/` folder. `cd istio && helmfile sync`
  * There is a bit of a gotcha here. The first time the helm chart is run, it installs the CustomResourceDefinitions used by Istio but not the actual pods/services. You need to run `helmfile sync` twice to install everything
4. Run `helmfile sync` in the `linkerd` folder to install the linkerd components.
5. As of this time, the linkerd helm chart doesn't do a good job of installing/integrating namerd so its a standalone kubernetes object. Install it by running `kubectl apply -f namerd.yaml`
### Exercise 1 - Basic Service Mesh
Now that the system components have been installed we can begin the demo. Starting with Istio, we'll install some services. This creates the three microservices that talk to each other. Its a very basic setup that mostly confirms that the mesh components are working correctly.
#### Istio
1. From the `istio` folder, run `kubectl apply -f 01-basic-mesh.yaml` to install the microservices.
2. `kubectl get pods -n istio-system` will show the progress. You should see the `words`, `simon`, and `capitalizer` pods scheduled with two containers each. The second container is the istio sidecar
3. Setup the variables for the IP/Port of the istio services: 
 * Ingress: `ISTIO_IP=$(minikube ip) && ISTIO_PORT=$(kubectl get service --namespace istio-system istio-ingress -o jsonpath='{.spec.ports[0].nodePort}')`
 * Grafana: `GRAFANA_PORT=$(kubectl get service --namespace istio-system istio-grafana -o jsonpath='{.spec.ports[0].nodePort}')`
 * Prometheus: `PROMETHEUS_PORT=$(kubectl get service --namespace istio-system istio-prometheus -o jsonpath='{.spec.ports[0].nodePort}')`
 * Zipkin: `ZIPKIN_PORT=$(kubectl get service --namespace istio-system istio-zipkin -o jsonpath='{.spec.ports[0].nodePort}')`
 * ServiceGraph: `SERVICEGRAPH_PORT=$(kubectl get service --namespace istio-system istio-servicegraph -o jsonpath='{.spec.ports[0].nodePort}')` 
4. Confirm services are working and talking to one another
 * `curl $ISTIO_IP:$ISTIO_PORT/word`
 * `curl $ISTIO_IP:$ISTIO_PORT/simon`
 * `curl $ISTIO_IP:$ISTIO_PORT/capitalize/word`
 * `curl $ISTIO_IP:$ISTIO_PORT/capitalize/simon`
 * Grafana: `open http://$ISTIO_IP:$GRAFANA_PORT`
 * Prometheus `open http://$ISTIO_IP:$PROMETHEUS_PORT`
 * ServiceGraph: `open http://$ISTIO_IP:$SERVICEGRAPH_PORT/dotviz`
 * Zipkin: `open http://$ISTIO_IP:$ZIPKIN_PORT/zipkin/`
#### Linkerd
1. From the `linkerd` folder run `kubectl apply -f 01-basic-mesh.yaml`
2. Setup the variables for the IP/Port of the linkerd service: `LINKERD_IP=$(minikube ip) && LINKERD_PORT=$(kubectl get service --namespace linkerd linkerd-linkerd -o jsonpath='{.spec.ports[0].nodePort}')  && LINKERD_UI_PORT=$(kubectl get service --namespace linkerd linkerd-linkerd -o jsonpath='{.spec.ports[2].nodePort}')`
3. Setup the variables for the namerd service: `NAMERD_IP=$(minikube ip) && NAMERD_API_PORT=$(kubectl get service --namespace linkerd namerd -o jsonpath='{.spec.ports[1].nodePort}') && NAMERD_UI_PORT=$(kubectl get service --namespace linkerd namerd -o jsonpath='{.spec.ports[2].nodePort}')`
4. Install the initial dtab with: `namerctl --base-url http://$NAMERD_IP:$NAMERD_API_PORT dtab create default 01-dtab-base`
5. Confirm services are working and talking to one another
 * `curl $LINKERD_IP:$LINKERD_PORT/word -H'Host: words'`
 * `curl $LINKERD_IP:$LINKERD_PORT/simon -H'Host: simon'`
 * `curl $LINKERD_IP:$LINKERD_PORT/capitalize/word -H'Host: capitalize'`
 * `curl $LINKERD_IP:$LINKERD_PORT/capitalize/simon -H'Host: capitalize'`
6. The Linkerd and Namerd consoles can be viewed with the following:
 * `open http://$LINKERD_IP:$LINKERD_UI_PORT`
 * `open http://$NAMERD_IP:$NAMERD_UI_PORT`
### Exercise 2 - Canary Deployment
This demonstrates a canary deployment of the words service. This means that two concurrent versions of the words service will be deployed. Traffic routing between the versions will be managed with the mesh. This will allow a small percentage of traffic to be sent to the new version to verify it is working as expected before shifting all traffic.

v1 returns words starting with "a". v2 returns words starting with "z"

#### Istio
1. Deploy words-v2 with `kubectl apply -f 02-canary-deploy.yaml`
2. See that there are two versions of the words service deployed but that only v1 ("a" words) is taking traffic: 
  * `kubectl get deployment -n istio-system -l app=words` 
  * `curl $ISTIO_IP:$ISTIO_PORT/word` should give a word starting with "a"
3. Check the other services are also getting the first version:
  * `curl $ISTIO_IP:$ISTIO_PORT/simon`
  * `curl $ISTIO_IP:$ISTIO_PORT/capitalize/word`
4. Move 10% of the traffic to v2 with: `kubectl apply -f 02a-canary-deploy.yaml`
5. You can observe the traffic with `watch -n1 curl -s $ISTIO_IP:$ISTIO_PORT/word` Roughly 10% of the traffic will be going to the new version
6. Move 50% traffic with `kubectl apply -f 02b-canary-deploy.yaml`
7. Finally shift all traffic with: `kubectl apply -f 02c-canary-deploy.yaml` Only "z" words should be returned.

#### Linkerd
1. Deploy words-v2 with `kubectl apply -f 02-canary-deploy.yaml && namerctl --base-url http://$NAMERD_IP:$NAMERD_API_PORT dtab update default 02a-canary-dtab`
2. See that there are two versions of the words service deployed but that only v1 ("a" words) is taking traffic: 
  * `kubectl get deployment -n linkerd -l app=words` 
  * `curl $LINKERD_IP:$LINKERD_PORT/word -H"Host: words"` should give a word starting with "a"
3. Check the other services are also getting the first version:
  * `curl $LINKERD_IP:$LINKERD_PORT/simon -H"Host: simon"`
  * `curl $LINKERD_IP:$LINKERD_PORT/capitalize/word -H"Host: capitalize"`
4. Move 10% of the traffic to v2 with: `namerctl --base-url http://$NAMERD_IP:$NAMERD_API_PORT dtab update default 02b-canary-dtab`
5. You can observe the traffic with `watch -n1 "curl -s  $LINKERD_IP:$LINKERD_PORT/word -H 'Host: words'"` Roughly 10% of the traffic will be going to the new version
6. Move 50% traffic with `namerctl --base-url http://$NAMERD_IP:$NAMERD_API_PORT dtab update default 02c-canary-dtab`
7. Finally shift all traffic with: `namerctl --base-url http://$NAMERD_IP:$NAMERD_API_PORT dtab update default 02d-canary-dtab` Only "z" words should be returned.

### Exercise 3 - Error Handling
This is similar to the canary deployment with the exception that v2 of the words service will throw a 500 error 50% of the time. With retry logic in the mesh, clients should be unaware of the error and the deployment can be rolled back safely.

#### Istio
1. Deploy words-v2 with `kubectl apply -f 03-errors-retry.yaml`
2. This starts with 10% of the traffic being sent to the bad v2. Curling the words service will show an "Internal Error" about 5% of the time.
4. Enable Retries with: `kubectl apply -f 03a-errors-retry.yaml`
5. The new version is still receiving 10% of the traffic and throwing errors. This time errors are retried automatically before a response is sent to the client.
6. Siege is a good tool to observe success/errors: `siege -c5 $ISTIO_IP:$ISTIO_PORT/word`

#### Linkerd
1. Retries are configured on the linkerd pod itself so we need to deploy a new version: `helmfile --file retry-charts.yaml sync
`
2. Restart the linkerd pod to pick up config changes: `kubectl delete pods -n linkerd -l app=linkerd-linkerd`
3. Install words-v2 `kubectl apply -f 03-errors-retry.yaml && namerctl --base-url http://$NAMERD_IP:$NAMERD_API_PORT dtab update default 03a-errors-dtab`
4. Traffic is split between the versions 50/50. Send traffic to a service and observe the linkerd console. Traffic to words-v2 will error 50% of the time but overall availability should still be 100%

### Exercise 4 - Stealth Deployment
The final exercise deploys two concurrent versions of the words service but only sends traffic to v2 if a specific header is passed along. This allows for validation/acceptance tests to be run against a live service in production without sending real traffic to the service. This method can be used in conjunction with the previous exercises to create a robust CD pipeline.

#### Istio
1. Deploy everything with `kubectl apply -f 04-stealh-deployment.yaml`
2. v1 of the service is accessed with `curl $ISTIO_IP:$ISTIO_PORT/word`
3. v2 of the service is accessed with the `X-Use-Canary` header: `curl $ISTIO_IP:$ISTIO_PORT/word -H'X-Use-Canary: true`

#### Linkerd
1. Deploy everything with `kubectl apply -f 04-stealh-deployment.yaml && namerctl --base-url http://$NAMERD_IP:$NAMERD_API_PORT dtab update default 04a-stealth-dtab`
2. v1 of the service is accessed with `curl $LINKERD_IP:$LINKERD_PORT/word -H"Host: words"`
3. v2 of the service is accessed with `curl $LINKERD_IP:$LINKERD_PORT/word -H"Host: words-canary"`