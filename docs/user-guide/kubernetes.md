# Kubernetes Ingress Controller

This guide explains how to use Træfik as an Ingress controller in a Kubernetes cluster.
If you are not familiar with Ingresses in Kubernetes you might want to read the [Kubernetes user guide](http://kubernetes.io/docs/user-guide/ingress/)

The config files used in this guide can be found in the [examples directory](https://github.com/containous/traefik/tree/master/examples/k8s)

## Prerequisites

1. A working Kubernetes cluster. If you want to follow along with this guide, you should setup [minikube](http://kubernetes.io/docs/getting-started-guides/minikube/)
on your machine, as it is the quickest way to get a local Kubernetes cluster setup for experimentation and development.

2. The `kubectl` binary should be [installed on your workstation](http://kubernetes.io/docs/getting-started-guides/minikube/#download-kubectl).

### Role Based Access Control configuration (Kubernetes 1.6+ only)

Kubernetes introduces [Role Based Access Control (RBAC)](https://kubernetes.io/docs/admin/authorization/rbac/) in 1.6+ to allow fine-grained control of Kubernetes resources and api.

If your cluster is configured with RBAC, you may need to authorize Træfik to use the Kubernetes API using ClusterRole and ClusterRoleBinding resources:

!!! note
    your cluster may have suitable ClusterRoles already setup, but the following should work everywhere

```yaml
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: traefik-ingress-controller
rules:
  - apiGroups:
      - ""
    resources:
      - pods
      - services
      - endpoints
      - secrets
    verbs:
      - get
      - list
      - watch
  - apiGroups:
      - extensions
    resources:
      - ingresses
    verbs:
      - get
      - list
      - watch
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: traefik-ingress-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: traefik-ingress-controller
subjects:
- kind: ServiceAccount
  name: traefik-ingress-controller
  namespace: kube-system
```

[examples/k8s/traefik-rbac.yaml](https://github.com/containous/traefik/tree/master/examples/k8s/traefik-rbac.yaml)

```shell
kubectl apply -f https://raw.githubusercontent.com/containous/traefik/master/examples/k8s/traefik-rbac.yaml
```

## Deploy Træfik using a Deployment or DaemonSet

It is possible to use Træfik with a [Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/) or a [DaemonSet](https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/) object,
 whereas both options have their own pros and cons:
 The scalability is much better when using a Deployment, because you will have a Single-Pod-per-Node model when using the DeaemonSet.
It is possible to exclusively run a Service on a dedicated set of machines using taints and tolerations with a DaemonSet.
On the other hand the DaemonSet allows you to access any Node directly on Port 80 and 443, where you have to setup a [Service](https://kubernetes.io/docs/concepts/services-networking/service/) object with a Deployment.

The Deployment objects looks like this:

```yml
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: traefik-ingress-controller
  namespace: kube-system
---
kind: Deployment
apiVersion: extensions/v1beta1
metadata:
  name: traefik-ingress-controller
  namespace: kube-system
  labels:
    k8s-app: traefik-ingress-lb
spec:
  replicas: 1
  selector:
    matchLabels:
      k8s-app: traefik-ingress-lb
  template:
    metadata:
      labels:
        k8s-app: traefik-ingress-lb
        name: traefik-ingress-lb
    spec:
      serviceAccountName: traefik-ingress-controller
      terminationGracePeriodSeconds: 60
      containers:
      - image: traefik
        name: traefik-ingress-lb
        args:
        - --web
        - --kubernetes
---
kind: Service
apiVersion: v1
metadata:
  name: traefik-ingress-service
spec:
  selector:
    k8s-app: traefik-ingress-lb
  ports:
    - protocol: TCP
      port: 80
    - protocol: TCP
      port: 8080
  type: NodePort
```
[examples/k8s/traefik-deployment.yaml](https://github.com/containous/traefik/tree/master/examples/k8s/traefik-deployment.yaml)

> The Service will expose two NodePorts which allow access to the ingress and the web interface.

The DaemonSet objects looks not much different:

```yaml
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: traefik-ingress-controller
  namespace: kube-system
---
kind: DaemonSet
apiVersion: extensions/v1beta1
metadata:
  name: traefik-ingress-controller
  namespace: kube-system
  labels:
    k8s-app: traefik-ingress-lb
spec:
  template:
    metadata:
      labels:
        k8s-app: traefik-ingress-lb
        name: traefik-ingress-lb
    spec:
      serviceAccountName: traefik-ingress-controller
      terminationGracePeriodSeconds: 60
      hostNetwork: true
      containers:
      - image: traefik
        name: traefik-ingress-lb
        ports:
        - name: http
          containerPort: 80
          hostPort: 80
        - name: admin
          containerPort: 8080
        securityContext:
          privileged: true
        args:
        - -d
        - --web
        - --kubernetes
---
kind: Service
apiVersion: v1
metadata:
  name: traefik-ingress-service
spec:
  selector:
    k8s-app: traefik-ingress-lb
  ports:
    - protocol: TCP
      port: 80
    - protocol: TCP
      port: 8080
  type: NodePort
```

[examples/k8s/traefik-ds.yaml](https://github.com/containous/traefik/tree/master/examples/k8s/traefik-ds.yaml)

To deploy Træfik to your cluster start by submitting one of the YAML files to the cluster with `kubectl`:

```shell
$ kubectl apply -f https://raw.githubusercontent.com/containous/traefik/master/examples/k8s/traefik-deployment.yaml
```

```shell
$ kubectl apply -f https://raw.githubusercontent.com/containous/traefik/master/examples/k8s/traefik-ds.yaml
```

There are some significant differences between using Deployments and DaemonSets.
The Deployment has easier up and down scaling possibilities. It can implement full pod lifecycle and supports rolling updates from Kubernetes 1.2.
At least one Pod is needed to run the Deployment.
The DaemonSet automatically scales to all nodes that meets a specific selector and guarantees to fill nodes one at a time.
Rolling updates are fully supported from Kubernetes 1.7 for DaemonSets as well.



### Check the Pods

Now lets check if our command was successful.

Start by listing the pods in the `kube-system` namespace:

```shell
$ kubectl --namespace=kube-system get pods

NAME                                         READY     STATUS    RESTARTS   AGE
kube-addon-manager-minikubevm                1/1       Running   0          4h
kubernetes-dashboard-s8krj                   1/1       Running   0          4h
traefik-ingress-controller-678226159-eqseo   1/1       Running   0          7m
```

You should see that after submitting the Deployment or DaemonSet to Kubernetes it has launched a Pod, and it is now running.
_It might take a few moments for kubernetes to pull the Træfik image and start the container._

> You could also check the deployment with the Kubernetes dashboard, run
> `minikube dashboard` to open it in your browser, then choose the `kube-system`
> namespace from the menu at the top right of the screen.

You should now be able to access Træfik on port 80 of your Minikube instance when using the DaemonSet:

```sh
curl $(minikube ip)
404 page not found
```

If you decided to use the deployment, then you need to target the correct NodePort, which can be seen then you execute `kubectl get services --namespace=kube-system`.

```sh
curl $(minikube ip):<NODEPORT>
404 page not found
```

> We expect to see a 404 response here as we haven't yet given Træfik any configuration.

## Deploy Træfik using Helm Chart

Instead of installing Træfik via an own object, you can also use the Træfik Helm chart.
This allows more complex configuration via Kubernetes [ConfigMap](https://kubernetes.io/docs/tasks/configure-pod-container/configmap/) and enabled TLS certificates.

Install Træfik chart by:

```shell
$ helm install stable/traefik
```

For more information, check out [the doc](https://github.com/kubernetes/charts/tree/master/stable/traefik).

## Submitting An Ingress to the cluster.

Lets start by creating a Service and an Ingress that will expose the [Træfik Web UI](https://github.com/containous/traefik#web-ui).

```yaml
apiVersion: v1
kind: Service
metadata:
  name: traefik-web-ui
  namespace: kube-system
spec:
  selector:
    k8s-app: traefik-ingress-lb
  ports:
  - port: 80
    targetPort: 8080
---
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: traefik-web-ui
  namespace: kube-system
  annotations:
    kubernetes.io/ingress.class: traefik
spec:
  rules:
  - host: traefik-ui.minikube
    http:
      paths:
      - backend:
          serviceName: traefik-web-ui
          servicePort: 80
```
[examples/k8s/ui.yaml](https://github.com/containous/traefik/tree/master/examples/k8s/ui.yaml)

```shell
kubectl apply -f https://raw.githubusercontent.com/containous/traefik/master/examples/k8s/ui.yaml
```

Now lets setup an entry in our /etc/hosts file to route `traefik-ui.minikube` to our cluster.

> In production you would want to set up real dns entries.

> You can get the ip address of your minikube instance by running `minikube ip`

```shell
echo "$(minikube ip) traefik-ui.minikube" | sudo tee -a /etc/hosts
```

We should now be able to visit [traefik-ui.minikube](http://traefik-ui.minikube) in the browser and view the Træfik Web UI.

## Name based routing

In this example we are going to setup websites for 3 of the United Kingdoms best loved cheeses, Cheddar, Stilton and Wensleydale.

First lets start by launching the 3 pods for the cheese websites.

```yaml
---
kind: Deployment
apiVersion: extensions/v1beta1
metadata:
  name: stilton
  labels:
    app: cheese
    cheese: stilton
spec:
  replicas: 2
  selector:
    matchLabels:
      app: cheese
      task: stilton
  template:
    metadata:
      labels:
        app: cheese
        task: stilton
        version: v0.0.1
    spec:
      containers:
      - name: cheese
        image: errm/cheese:stilton
        resources:
          requests:
            cpu: 100m
            memory: 50Mi
          limits:
            cpu: 100m
            memory: 50Mi
        ports:
        - containerPort: 80
---
kind: Deployment
apiVersion: extensions/v1beta1
metadata:
  name: cheddar
  labels:
    app: cheese
    cheese: cheddar
spec:
  replicas: 2
  selector:
    matchLabels:
      app: cheese
      task: cheddar
  template:
    metadata:
      labels:
        app: cheese
        task: cheddar
        version: v0.0.1
    spec:
      containers:
      - name: cheese
        image: errm/cheese:cheddar
        resources:
          requests:
            cpu: 100m
            memory: 50Mi
          limits:
            cpu: 100m
            memory: 50Mi
        ports:
        - containerPort: 80
---
kind: Deployment
apiVersion: extensions/v1beta1
metadata:
  name: wensleydale
  labels:
    app: cheese
    cheese: wensleydale
spec:
  replicas: 2
  selector:
    matchLabels:
      app: cheese
      task: wensleydale
  template:
    metadata:
      labels:
        app: cheese
        task: wensleydale
        version: v0.0.1
    spec:
      containers:
      - name: cheese
        image: errm/cheese:wensleydale
        resources:
          requests:
            cpu: 100m
            memory: 50Mi
          limits:
            cpu: 100m
            memory: 50Mi
        ports:
        - containerPort: 80
```
[examples/k8s/cheese-deployments.yaml](https://github.com/containous/traefik/tree/master/examples/k8s/cheese-deployments.yaml)

```shell
kubectl apply -f https://raw.githubusercontent.com/containous/traefik/master/examples/k8s/cheese-deployments.yaml
```

Next we need to setup a service for each of the cheese pods.

```yaml
---
apiVersion: v1
kind: Service
metadata:
  name: stilton
spec:
  ports:
  - name: http
    targetPort: 80
    port: 80
  selector:
    app: cheese
    task: stilton
---
apiVersion: v1
kind: Service
metadata:
  name: cheddar
spec:
  ports:
  - name: http
    targetPort: 80
    port: 80
  selector:
    app: cheese
    task: cheddar
---
apiVersion: v1
kind: Service
metadata:
  name: wensleydale
  annotations:
    traefik.backend.circuitbreaker: "NetworkErrorRatio() > 0.5"
spec:
  ports:
  - name: http
    targetPort: 80
    port: 80
  selector:
    app: cheese
    task: wensleydale
```

> Notice that we also set a [circuit breaker expression](https://docs.traefik.io/basics/#backends) for one of the backends
> by setting the `traefik.backend.circuitbreaker` annotation on the service.


[examples/k8s/cheese-services.yaml](https://github.com/containous/traefik/tree/master/examples/k8s/cheese-services.yaml)

```shell
kubectl apply -f https://raw.githubusercontent.com/containous/traefik/master/examples/k8s/cheese-services.yaml
```

Now we can submit an ingress for the cheese websites.

```yaml
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: cheese
  annotations:
    kubernetes.io/ingress.class: traefik
spec:
  rules:
  - host: stilton.minikube
    http:
      paths:
      - path: /
        backend:
          serviceName: stilton
          servicePort: http
  - host: cheddar.minikube
    http:
      paths:
      - path: /
        backend:
          serviceName: cheddar
          servicePort: http
  - host: wensleydale.minikube
    http:
      paths:
      - path: /
        backend:
          serviceName: wensleydale
          servicePort: http
```
[examples/k8s/cheese-ingress.yaml](https://github.com/containous/traefik/tree/master/examples/k8s/cheese-ingress.yaml)

> Notice that we list each hostname, and add a backend service.

```shell
kubectl apply -f https://raw.githubusercontent.com/containous/traefik/master/examples/k8s/cheese-ingress.yaml
```

Now visit the [Træfik dashboard](http://traefik-ui.minikube/) and you should see a frontend for each host. Along with a backend listing for each service with a Server set up for each pod.

If you edit your `/etc/hosts` again you should be able to access the cheese websites in your browser.

```shell
echo "$(minikube ip) stilton.minikube cheddar.minikube wensleydale.minikube" | sudo tee -a /etc/hosts
```

* [Stilton](http://stilton.minikube/)
* [Cheddar](http://cheddar.minikube/)
* [Wensleydale](http://wensleydale.minikube/)

## Path based routing

Now lets suppose that our fictional client has decided that while they are super happy about our cheesy web design, when they asked for 3 websites they had not really bargained on having to buy 3 domain names.

No problem, we say, why don't we reconfigure the sites to host all 3 under one domain.


```yaml
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: cheeses
  annotations:
    kubernetes.io/ingress.class: traefik
    traefik.frontend.rule.type: PathPrefixStrip
spec:
  rules:
  - host: cheeses.minikube
    http:
      paths:
      - path: /stilton
        backend:
          serviceName: stilton
          servicePort: http
      - path: /cheddar
        backend:
          serviceName: cheddar
          servicePort: http
      - path: /wensleydale
        backend:
          serviceName: wensleydale
          servicePort: http
```
[examples/k8s/cheeses-ingress.yaml](https://github.com/containous/traefik/tree/master/examples/k8s/cheeses-ingress.yaml)

> Notice that we are configuring Træfik to strip the prefix from the url path
> with the `traefik.frontend.rule.type` annotation so that we can use
> the containers from the previous example without modification.

```shell
kubectl apply -f https://raw.githubusercontent.com/containous/traefik/master/examples/k8s/cheeses-ingress.yaml
```

```shell
echo "$(minikube ip) cheeses.minikube" | sudo tee -a /etc/hosts
```

You should now be able to visit the websites in your browser.

* [cheeses.minikube/stilton](http://cheeses.minikube/stilton/)
* [cheeses.minikube/cheddar](http://cheeses.minikube/cheddar/)
* [cheeses.minikube/wensleydale](http://cheeses.minikube/wensleydale/)

## Specifying priority for routing

Sometimes you need to specify priority for ingress route, especially when handling wildcard routes.
This can be done by adding annotation `traefik.frontend.priority`, i.e.:

```yaml
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: wildcard-cheeses
  annotations:
    traefik.frontend.priority: 1
spec:
  rules:
  - host: *.minikube
    http:
      paths:
      - path: /
        backend:
          serviceName: stilton
          servicePort: http

kind: Ingress
metadata:
  name: specific-cheeses
  annotations:
    traefik.frontend.priority: 2
spec:
  rules:
  - host: specific.minikube
    http:
      paths:
      - path: /
        backend:
          serviceName: stilton
          servicePort: http
```


## Forwarding to ExternalNames

When specifying an [ExternalName](https://kubernetes.io/docs/concepts/services-networking/service/#services-without-selectors),
Træfik will forward requests to the given host accordingly and use HTTPS when the Service port matches 443.
This still requires setting up a proper port mapping on the Service from the Ingress port to the (external) Service port.

## Disable passing the Host header

By default Træfik will pass the incoming Host header on to the upstream resource.
There are times however where you may not want this to be the case.
For example if your service is of the ExternalName type.

### Disable entirely

Add the following to your toml config:
```toml
disablePassHostHeaders = true
```

### Disable per ingress

To disable passing the Host header per ingress resource set the `traefik.frontend.passHostHeader` annotation on your ingress to `false`.

Here is an example ingress definition:
```yaml
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: example
  annotations:
    kubernetes.io/ingress.class: traefik
    traefik.frontend.passHostHeader: "false"
spec:
  rules:
  - host: example.com
    http:
      paths:
      - path: /static
        backend:
          serviceName: static
          servicePort: https
```

And an example service definition:
```yaml
apiVersion: v1
kind: Service
metadata:
  name: static
spec:
  ports:
  - name: https
    port: 443
  type: ExternalName
  externalName: static.otherdomain.com
```

If you were to visit `example.com/static` the request would then be passed onto `static.otherdomain.com/static` and s`tatic.otherdomain.com` would receive the request with the Host header being `static.otherdomain.com`.

!!! note
    The per ingress annotation overides whatever the global value is set to.
    So you could set `disablePassHostHeaders` to `true` in your toml file and then enable passing 
    the host header per ingress if you wanted.

## Excluding an ingress from Træfik

You can control which ingress Træfik cares about by using the `kubernetes.io/ingress.class` annotation.
By default if the annotation is not set at all Træfik will include the ingress.
If the annotation is set to anything other than traefik or a blank string Træfik will ignore it.


![](http://i.giphy.com/ujUdrdpX7Ok5W.gif)
