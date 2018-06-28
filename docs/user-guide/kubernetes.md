# Kubernetes Ingress Controller

This guide explains how to use Træfik as an Ingress controller for a Kubernetes cluster.

If you are not familiar with Ingresses in Kubernetes you might want to read the [Kubernetes user guide](https://kubernetes.io/docs/concepts/services-networking/ingress/)

The config files used in this guide can be found in the [examples directory](https://github.com/containous/traefik/tree/master/examples/k8s)

## Prerequisites

1. A working Kubernetes cluster. If you want to follow along with this guide, you should setup [minikube](https://kubernetes.io/docs/getting-started-guides/minikube/) on your machine, as it is the quickest way to get a local Kubernetes cluster setup for experimentation and development.

!!! note
    The guide is likely not fully adequate for a production-ready setup.

2. The `kubectl` binary should be [installed on your workstation](https://kubernetes.io/docs/getting-started-guides/minikube/#download-kubectl).

### Role Based Access Control configuration (Kubernetes 1.6+ only)

Kubernetes introduces [Role Based Access Control (RBAC)](https://kubernetes.io/docs/admin/authorization/rbac/) in 1.6+ to allow fine-grained control of Kubernetes resources and API.

If your cluster is configured with RBAC, you will need to authorize Træfik to use the Kubernetes API. There are two ways to set up the proper permission: Via namespace-specific RoleBindings or a single, global ClusterRoleBinding.

RoleBindings per namespace enable to restrict granted permissions to the very namespaces only that Træfik is watching over, thereby following the least-privileges principle. This is the preferred approach if Træfik is not supposed to watch all namespaces, and the set of namespaces does not change dynamically. Otherwise, a single ClusterRoleBinding must be employed.

!!! note
    RoleBindings per namespace are available in Træfik 1.5 and later. Please use ClusterRoleBindings for older versions.

For the sake of simplicity, this guide will use a ClusterRoleBinding:

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

For namespaced restrictions, one RoleBinding is required per watched namespace along with a corresponding configuration of Træfik's `kubernetes.namespaces` parameter.

## Deploy Træfik using a Deployment or DaemonSet

It is possible to use Træfik with a [Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/) or a [DaemonSet](https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/) object,
 whereas both options have their own pros and cons:

- The scalability can be much better when using a Deployment, because you will have a Single-Pod-per-Node model when using a DaemonSet, whereas you may need less replicas based on your environment when using a Deployment.
- DaemonSets automatically scale to new nodes, when the nodes join the cluster, whereas Deployment pods are only scheduled on new nodes if required.
- DaemonSets ensure that only one replica of pods run on any single node. Deployments require affinity settings if you want to ensure that two pods don't end up on the same node.
- DaemonSets can be run with the `NET_BIND_SERVICE` capability, which will allow it to bind to port 80/443/etc on each host. This will allow bypassing the kube-proxy, and reduce traffic hops. Note that this is against the Kubernetes Best Practices [Guidelines](https://kubernetes.io/docs/concepts/configuration/overview/#services), and raises the potential for scheduling/scaling issues. Despite potential issues, this remains the choice for most ingress controllers.
- If you are unsure which to choose, start with the Daemonset.

The Deployment objects looks like this:

```yaml
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
        ports:
        - name: http
          containerPort: 80
        - name: admin
          containerPort: 8080
        args:
        - --api
        - --kubernetes
        - --logLevel=INFO
---
kind: Service
apiVersion: v1
metadata:
  name: traefik-ingress-service
  namespace: kube-system
spec:
  selector:
    k8s-app: traefik-ingress-lb
  ports:
    - protocol: TCP
      port: 80
      name: web
    - protocol: TCP
      port: 8080
      name: admin
  type: NodePort
```

[examples/k8s/traefik-deployment.yaml](https://github.com/containous/traefik/tree/master/examples/k8s/traefik-deployment.yaml)

!!! note
    The Service will expose two NodePorts which allow access to the ingress and the web interface.

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
          capabilities:
            drop:
            - ALL
            add:
            - NET_BIND_SERVICE
        args:
        - --api
        - --kubernetes
        - --logLevel=INFO
---
kind: Service
apiVersion: v1
metadata:
  name: traefik-ingress-service
  namespace: kube-system
spec:
  selector:
    k8s-app: traefik-ingress-lb
  ports:
    - protocol: TCP
      port: 80
      name: web
    - protocol: TCP
      port: 8080
      name: admin
```

[examples/k8s/traefik-ds.yaml](https://github.com/containous/traefik/tree/master/examples/k8s/traefik-ds.yaml)

!!! note
    This will create a Daemonset that uses privileged ports 80/8080 on the host. This may not work on all providers, but illustrates the static (non-NodePort) hostPort binding. The `traefik-ingress-service` can still be used inside the cluster to access the DaemonSet pods.

To deploy Træfik to your cluster start by submitting one of the YAML files to the cluster with `kubectl`:

```shell
kubectl apply -f https://raw.githubusercontent.com/containous/traefik/master/examples/k8s/traefik-deployment.yaml
```

```shell
kubectl apply -f https://raw.githubusercontent.com/containous/traefik/master/examples/k8s/traefik-ds.yaml
```

There are some significant differences between using Deployments and DaemonSets:

- The Deployment has easier up and down scaling possibilities.
    It can implement full pod lifecycle and supports rolling updates from Kubernetes 1.2.
    At least one Pod is needed to run the Deployment.
- The DaemonSet automatically scales to all nodes that meets a specific selector and guarantees to fill nodes one at a time.
    Rolling updates are fully supported from Kubernetes 1.7 for DaemonSets as well.

### Check the Pods

Now lets check if our command was successful.

Start by listing the pods in the `kube-system` namespace:

```shell
kubectl --namespace=kube-system get pods
```

```shell
NAME                                         READY     STATUS    RESTARTS   AGE
kube-addon-manager-minikubevm                1/1       Running   0          4h
kubernetes-dashboard-s8krj                   1/1       Running   0          4h
traefik-ingress-controller-678226159-eqseo   1/1       Running   0          7m
```

You should see that after submitting the Deployment or DaemonSet to Kubernetes it has launched a Pod, and it is now running.
_It might take a few moments for Kubernetes to pull the Træfik image and start the container._

!!! note
    You could also check the deployment with the Kubernetes dashboard, run
    `minikube dashboard` to open it in your browser, then choose the `kube-system`
    namespace from the menu at the top right of the screen.

You should now be able to access Træfik on port 80 of your Minikube instance when using the DaemonSet:

```shell
curl $(minikube ip)
```

```shell
404 page not found
```

If you decided to use the deployment, then you need to target the correct NodePort, which can be seen when you execute `kubectl get services --namespace=kube-system`.

```shell
curl $(minikube ip):<NODEPORT>
```

```shell
404 page not found
```

!!! note
    We expect to see a 404 response here as we haven't yet given Træfik any configuration.

All further examples below assume a DaemonSet installation. Deployment users will need to append the NodePort when constructing requests.

## Deploy Træfik using Helm Chart

!!! note
    The Helm Chart is maintained by the community, not the Træfik project maintainers.

Instead of installing Træfik via Kubernetes object directly, you can also use the Træfik Helm chart.

Install the Træfik chart by:

```shell
helm install stable/traefik
```
Install the Træfik chart using a values.yaml file.

```shell
helm install --values values.yaml stable/traefik
```

```yaml
dashboard:
  enabled: true
  domain: traefik-ui.minikube
kubernetes:
  namespaces:
    - default
    - kube-system
```
For more information, check out [the documentation](https://github.com/kubernetes/charts/tree/master/stable/traefik).

## Submitting an Ingress to the Cluster

Lets start by creating a Service and an Ingress that will expose the [Træfik Web UI](https://github.com/containous/traefik#web-ui).

```yaml
---
apiVersion: v1
kind: Service
metadata:
  name: traefik-web-ui
  namespace: kube-system
spec:
  selector:
    k8s-app: traefik-ingress-lb
  ports:
  - name: web
    port: 80
    targetPort: 8080
---
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: traefik-web-ui
  namespace: kube-system
spec:
  rules:
  - host: traefik-ui.minikube
    http:
      paths:
      - path: /
        backend:
          serviceName: traefik-web-ui
servicePort: web
```

[examples/k8s/ui.yaml](https://github.com/containous/traefik/tree/master/examples/k8s/ui.yaml)

```shell
kubectl apply -f https://raw.githubusercontent.com/containous/traefik/master/examples/k8s/ui.yaml
```

Now lets setup an entry in our `/etc/hosts` file to route `traefik-ui.minikube` to our cluster.

In production you would want to set up real DNS entries.
You can get the IP address of your minikube instance by running `minikube ip`:

```shell
echo "$(minikube ip) traefik-ui.minikube" | sudo tee -a /etc/hosts
```

We should now be able to visit [traefik-ui.minikube](http://traefik-ui.minikube) in the browser and view the Træfik web UI.

### Add a TLS Certificate to the Ingress

!!! note
    For this example to work you need a TLS entrypoint. You don't have to provide a TLS certificate at this point.
    For more details see [here](/configuration/entrypoints/).

To setup an HTTPS-protected ingress, you can leverage the TLS feature of the ingress resource.

```yaml
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
  tls:
   - secretName: traefik-ui-tls-cert
```

In addition to the modified ingress you need to provide the TLS certificate via a Kubernetes secret in the same namespace as the ingress.
The following two commands will generate a new certificate and create a secret containing the key and cert files.

```shell
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout tls.key -out tls.crt -subj "/CN=traefik-ui.minikube"
kubectl -n kube-system create secret tls traefik-ui-tls-cert --key=tls.key --cert=tls.crt
```

If there are any errors while loading the TLS section of an ingress, the whole ingress will be skipped.

!!! note
    The secret must have two entries named `tls.key`and `tls.crt`.
    See the [Kubernetes documentation](https://kubernetes.io/docs/concepts/services-networking/ingress/#tls) for more details.

!!! note
    The TLS certificates will be added to all entrypoints defined by the ingress annotation `traefik.frontend.entryPoints`.
    If no such annotation is provided, the TLS certificates will be added to all TLS-enabled `defaultEntryPoints`.

!!! note
    The field `hosts` in the TLS configuration is ignored. Instead, the domains provided by the certificate are used for this purpose.
    It is recommended to not use wildcard certificates as they will match globally.

## Basic Authentication

It's possible to protect access to Træfik through basic authentication. (See the [Kubernetes Ingress](/configuration/backends/kubernetes) configuration page for syntactical details and restrictions.)

### Creating the Secret

A. Use `htpasswd` to create a file containing the username and the MD5-encoded password:

```shell
htpasswd -c ./auth myusername
```

You will be prompted for a password which you will have to enter twice.
`htpasswd` will create a file with the following:

```shell
cat auth
```

```shell
myusername:$apr1$78Jyn/1K$ERHKVRPPlzAX8eBtLuvRZ0
```

B. Now use `kubectl` to create a secret in the `monitoring` namespace using the file created by `htpasswd`.

```shell
kubectl create secret generic mysecret --from-file auth --namespace=monitoring
```

!!! note
    Secret must be in same namespace as the Ingress object.

C. Attach the following annotations to the Ingress object:

- `ingress.kubernetes.io/auth-type: "basic"`
- `ingress.kubernetes.io/auth-secret: "mysecret"`

They specify basic authentication and reference the Secret `mysecret` containing the credentials.

Following is a full Ingress example based on Prometheus:

```yaml
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
 name: prometheus-dashboard
 namespace: monitoring
 annotations:
   kubernetes.io/ingress.class: traefik
   ingress.kubernetes.io/auth-type: "basic"
   ingress.kubernetes.io/auth-secret: "mysecret"
spec:
 rules:
 - host: dashboard.prometheus.example.com
   http:
     paths:
     - backend:
         serviceName: prometheus
         servicePort: 9090
```

You can apply the example as following:

```shell
kubectl create -f prometheus-ingress.yaml -n monitoring
```

## Name-based Routing

In this example we are going to setup websites for three of the United Kingdoms best loved cheeses: Cheddar, Stilton, and Wensleydale.

First lets start by launching the pods for the cheese websites.

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
        ports:
        - containerPort: 80
```

[examples/k8s/cheese-deployments.yaml](https://github.com/containous/traefik/tree/master/examples/k8s/cheese-deployments.yaml)

```shell
kubectl apply -f https://raw.githubusercontent.com/containous/traefik/master/examples/k8s/cheese-deployments.yaml
```

Next we need to setup a Service for each of the cheese pods.

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

!!! note
    We also set a [circuit breaker expression](/basics/#backends) for one of the backends by setting the `traefik.backend.circuitbreaker` annotation on the service.

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

!!! note
    we list each hostname, and add a backend service.

```shell
kubectl apply -f https://raw.githubusercontent.com/containous/traefik/master/examples/k8s/cheese-ingress.yaml
```

Now visit the [Træfik dashboard](http://traefik-ui.minikube/) and you should see a frontend for each host.
Along with a backend listing for each service with a server set up for each pod.

If you edit your `/etc/hosts` again you should be able to access the cheese websites in your browser.

```shell
echo "$(minikube ip) stilton.minikube cheddar.minikube wensleydale.minikube" | sudo tee -a /etc/hosts
```

- [Stilton](http://stilton.minikube/)
- [Cheddar](http://cheddar.minikube/)
- [Wensleydale](http://wensleydale.minikube/)

## Path-based Routing

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

!!! note
    We are configuring Træfik to strip the prefix from the url path with the `traefik.frontend.rule.type` annotation so that we can use the containers from the previous example without modification.

```shell
kubectl apply -f https://raw.githubusercontent.com/containous/traefik/master/examples/k8s/cheeses-ingress.yaml
```

```shell
echo "$(minikube ip) cheeses.minikube" | sudo tee -a /etc/hosts
```

You should now be able to visit the websites in your browser.

- [cheeses.minikube/stilton](http://cheeses.minikube/stilton/)
- [cheeses.minikube/cheddar](http://cheeses.minikube/cheddar/)
- [cheeses.minikube/wensleydale](http://cheeses.minikube/wensleydale/)

## Specifying Routing Priorities

Sometimes you need to specify priority for ingress routes, especially when handling wildcard routes.
This can be done by adding the `traefik.frontend.priority` annotation, i.e.:

```yaml
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: wildcard-cheeses
  annotations:
    traefik.frontend.priority: "1"
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
    traefik.frontend.priority: "2"
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

Note that priority values must be quoted to avoid numeric interpretation (which are illegal for annotations).

## Forwarding to ExternalNames

When specifying an [ExternalName](https://kubernetes.io/docs/concepts/services-networking/service/#services-without-selectors),
Træfik will forward requests to the given host accordingly and use HTTPS when the Service port matches 443.
This still requires setting up a proper port mapping on the Service from the Ingress port to the (external) Service port.

## Disable passing the Host Header

By default Træfik will pass the incoming Host header to the upstream resource.

However, there are times when you may not want this to be the case. For example, if your service is of the ExternalName type.

### Disable globally

Add the following to your TOML configuration file:

```toml
disablePassHostHeaders = true
```

### Disable per Ingress

To disable passing the Host header per ingress resource set the `traefik.frontend.passHostHeader` annotation on your ingress to `"false"`.

Here is an example definition:

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

If you were to visit `example.com/static` the request would then be passed on to `static.otherdomain.com/static`, and `static.otherdomain.com` would receive the request with the Host header being `static.otherdomain.com`.

!!! note
    The per-ingress annotation overrides whatever the global value is set to.
    So you could set `disablePassHostHeaders` to `true` in your TOML configuration file and then enable passing the host header per ingress if you wanted.

## Partitioning the Ingress object space

By default, Træfik processes every Ingress objects it observes. At times, however, it may be desirable to ignore certain objects. The following sub-sections describe common use cases and how they can be handled with Træfik.

### Between Træfik and other Ingress controller implementations

Sometimes Træfik runs along other Ingress controller implementations. One such example is when both Træfik and a cloud provider Ingress controller are active.

The `kubernetes.io/ingress.class` annotation can be attached to any Ingress object in order to control whether Træfik should handle it.

If the annotation is missing, contains an empty value, or the value `traefik`, then the Træfik controller will take responsibility and process the associated Ingress object.

It is also possible to set the `ingressClass` option in Træfik to a particular value. Træfik will only process matching Ingress objects.
For instance, setting the option to `traefik-internal` causes Træfik to process Ingress objects with the same `kubernetes.io/ingress.class` annotation value, ignoring all other objects (including those with a `traefik` value, empty value, and missing annotation).

!!! note
    Letting multiple ingress controllers handle the same ingress objects can lead to unintended behavior.
    It is recommended to prefix all ingressClass values with `traefik` to avoid unintended collisions with other ingress implementations.

### Between multiple Træfik Deployments

Sometimes multiple Træfik Deployments are supposed to run concurrently.
For instance, it is conceivable to have one Deployment deal with internal and another one with external traffic.

For such cases, it is advisable to classify Ingress objects through a label and configure the `labelSelector` option per each Træfik Deployment accordingly.
To stick with the internal/external example above, all Ingress objects meant for internal traffic could receive a `traffic-type: internal` label while objects designated for external traffic receive a `traffic-type: external` label.
The label selectors on the Træfik Deployments would then be `traffic-type=internal` and `traffic-type=external`, respectively.

## Production advice

### Resource limitations

The examples shown deliberately do not specify any [resource limitations](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/) as there is no one size fits all.

In a production environment, however, it is important to set proper bounds, especially with regards to CPU:

- too strict and Træfik will be throttled while serving requests (as Kubernetes imposes hard quotas)
- too loose and Træfik may waste resources not available for other containers

When in doubt, you should measure your resource needs, and adjust requests and limits accordingly.
