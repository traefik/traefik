---
title: "Migrate from NGINX Ingress Controller to Traefik"
description: "Step-by-step guide to migrate from Kubernetes NGINX Ingress Controller to Traefik with minimal downtime and annotation compatibility."
---

# Migrate from NGINX Ingress Controller to Traefik

How to migrate from NGINX Ingress Controller to Traefik with minimal downtime.
{: .subtitle }

## Introduction

This guide provides a step-by-step migration path from the Kubernetes NGINX Ingress Controller to Traefik. By leveraging Traefik Kubernetes Ingress NGINX Provider, you can migrate your workloads without rewriting your Ingress configurations, ensuring a smooth transition with minimal downtime.

---

!!! danger "NGINX Ingress Controller Retirement"

    The Kubernetes NGINX Ingress Controller project has announced its retirement in **March 2026**. After this date:

    - No new releases or updates
    - No security patches
    - No bug fixes

    For more information, see the [official Kubernetes blog announcement](https://kubernetes.io/blog/2025/11/11/ingress-nginx-retirement).

!!! info "Traefik Version Requirement"

    The Kubernetes Ingress NGINX provider requires **Traefik v3.6.2 or later**. Ensure you are using a compatible version before proceeding with this migration.

---

## Prerequisites

Before starting the migration, ensure you have:

- **Existing NGINX Ingress Controller** running in your Kubernetes cluster
- **Kubernetes cluster access** with `kubectl` configured
- **Helm** 
- **Cluster admin permissions** to create RBAC resources
- **Backup of critical configurations** (Ingress resources, ConfigMaps, Secrets)
- **Traefik v3.6.2 or later** (the Kubernetes Ingress NGINX provider requires this version)

!!! tip "Backup Recommendations"

    ```bash
    # Export all Ingress resources
    kubectl get ingress --all-namespaces -o yaml > ingress-backup.yaml

    # Export NGINX ConfigMaps
    kubectl get configmap --all-namespaces -l app.kubernetes.io/name=ingress-nginx -o yaml > nginx-configmaps.yaml
    ```

---

## Migration Strategy Overview

This migration follows a **quick switchover approach** with a focus on **load balancer IP retention** to minimize downtime and DNS propagation delays.

**Migration Flow:**

1. Prepare Traefik Helm values configuration
2. Delete NGINX LoadBalancer service to release the IP (brief downtime starts)
3. Uninstall NGINX Ingress Controller
4. Install Traefik with the released IP and IngressClass configuration
5. Verify traffic routing through Traefik (downtime ends)
6. Monitor and validate

**Expected Downtime:** Brief downtime (typically seconds to a few minutes) during the switchover, depending on cloud provider IP release timing. If using DNS-based migration instead of IP retention, downtime depends on DNS TTL.

---

## Step 1: Install Traefik with Helm

Traefik can be installed using Helm, which automatically creates all necessary RBAC resources and configures the Kubernetes Ingress NGINX provider.

### Add Traefik Helm Repository

```bash
# Add the Traefik Helm repository
helm repo add traefik https://traefik.github.io/charts

# Update your local Helm chart repository cache
helm repo update
```

### Prepare Helm Values Configuration

Before installing, identify your existing NGINX Ingress Controller's Load Balancer IP to retain it:

```bash
# Get the current Load Balancer IP from NGINX service
kubectl get svc -n ingress-nginx ingress-nginx-controller -o jsonpath='{.status.loadBalancer.ingress[0].ip}'

# Or for hostname-based load balancers (AWS ELB)
kubectl get svc -n ingress-nginx ingress-nginx-controller -o jsonpath='{.status.loadBalancer.ingress[0].hostname}'
```

Create a `traefik-values.yaml` file with the following configuration:

```yaml tab="Helm Chart Values"
# Enable the Kubernetes Ingress NGINX provider
providers:
  kubernetesIngressNginx:
    enabled: true
    # IngressClass configuration
    ingressClass: "nginx"
    controllerClass: "k8s.io/ingress-nginx"
    # Watch Ingresses without IngressClass (optional, use with caution)
    watchIngressWithoutClass: false
    # Publish the Traefik service address back to Ingress status
    publishService:
      enabled: true

# Configure EntryPoints
ports:
  web:
    port: 8000
    expose:
      default: true
    exposedPort: 80
    protocol: TCP
  websecure:
    port: 8443
    expose:
      default: true
    exposedPort: 443
    protocol: TCP
    tls:
      enabled: true

# Service configuration (see Load Balancer IP Retention section)
service:
  enabled: true
  type: LoadBalancer
  # Additional cloud-provider specific configuration
  annotations: {}
  spec: {}
    # Configure loadBalancerIP here (see Step 2)
    # loadBalancerIP: "<your-existing-ip>"

# Create the nginx IngressClass (required after removing NGINX Ingress Controller)
# This ensures Traefik continues to recognize Ingresses with ingressClassName: nginx
extraObjects:
  - apiVersion: networking.k8s.io/v1
    kind: IngressClass
    metadata:
      name: nginx
    spec:
      controller: k8s.io/ingress-nginx
```

!!! info "RBAC Included"

    The Traefik Helm chart automatically creates all necessary RBAC resources:

    - ServiceAccount
    - ClusterRole with permissions for Ingresses, Services, Endpoints, Secrets
    - ClusterRoleBinding

    No manual RBAC configuration is required.

!!! warning "IngressClass Preservation"

    The `extraObjects` section creates the `nginx` IngressClass that Traefik needs to recognize your existing Ingresses. When you uninstall the NGINX Ingress Controller via Helm, it deletes the `nginx` IngressClass. Without this IngressClass, Traefik cannot route traffic to Ingresses that specify `ingressClassName: nginx`.

    **Important:** If NGINX is still installed, the IngressClass already exists and Helm will report a conflict. You must uninstall NGINX **before** installing Traefik with the `extraObjects` configuration, or install Traefik without `extraObjects` first and manually create the IngressClass after uninstalling NGINX.

---

## Step 2: Load Balancer IP Retention

To minimize downtime during migration, it's critical to retain your existing Load Balancer IP address. This prevents DNS propagation delays and ensures continuous service availability.

### General Approach

The `loadBalancerIP` field allows you to specify a static IP for the LoadBalancer service. However, cloud providers handle this differently.

!!! warning "Cloud Provider Variations"

    Load balancer IP retention strategies vary by cloud provider. Follow the appropriate section for your environment.

!!! warning "ServiceLB and hostPort-based environments"

    On k3s/k3d (and any environment that implements `Service.Type=LoadBalancer` by binding host ports such as Klipper/ServiceLB), only one LoadBalancer service can listen on ports 80/443 at a time. Delete the NGINX LoadBalancer service before creating Traefik's service, or use different ports, otherwise Traefik's service will stay in `Pending`.

### AWS Elastic Load Balancer

AWS does not support static IPs for Classic Load Balancers (CLB). Instead, use **Network Load Balancers (NLB)** with Elastic IPs.

**Option A: Pre-allocate Elastic IPs**

```bash
# Allocate Elastic IPs (one per availability zone)
aws ec2 allocate-address --domain vpc --region us-east-1

# Note the AllocationId for each EIP
```

Update `traefik-values.yaml`:

```yaml
service:
  type: LoadBalancer
  loadBalancerClass: service.k8s.aws/nlb  # uses AWS Load Balancer Controller
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-type: "external"
    service.beta.kubernetes.io/aws-load-balancer-nlb-target-type: "ip"
    # Optional: ensure internet-facing scheme (default is internet-facing)
    # service.beta.kubernetes.io/aws-load-balancer-scheme: "internet-facing"
    service.beta.kubernetes.io/aws-load-balancer-eip-allocations: "eipalloc-xxxxxxxxx,eipalloc-yyyyyyyyy"
  # Optional (for IP target mode): disable NodePort allocation
  # spec:
  #   allocateLoadBalancerNodePorts: false
```

**Option B: Migrate DNS to New Load Balancer**

If retaining the exact IP isn't possible, update your DNS records to point to the new NLB hostname:

1. Deploy Traefik with a new Load Balancer
2. Update DNS CNAME records to point to the new NLB
3. Wait for DNS propagation
4. [Decommission NGINX](#step-7-decommission-nginx-ingress-controller)

### Azure Load Balancer

Azure supports static public IPs for Load Balancers. You can reuse the existing public IP resource.

**Identify Existing Public IP:**

```bash
# Get the public IP resource name
az network public-ip list --resource-group <your-resource-group> --query "[?ipAddress=='<your-ip>'].name" -o tsv
```

Update `traefik-values.yaml`:

```yaml
service:
  type: LoadBalancer
  annotations:
    service.beta.kubernetes.io/azure-load-balancer-resource-group: "<your-resource-group>"
  spec:
    loadBalancerIP: "<your-existing-ip>"
```

!!! tip "Deleting NGINX Service"

    To successfully reuse the IP, you must delete the NGINX service **before** creating the Traefik service:

    ```bash
    # Scale down NGINX first
    kubectl scale deployment -n ingress-nginx ingress-nginx-controller --replicas=0

    # Delete NGINX service to release the IP
    kubectl delete svc -n ingress-nginx ingress-nginx-controller

    # Install Traefik immediately after
    helm install traefik traefik/traefik -f traefik-values.yaml
    ```

### GCP Load Balancer

GCP supports static IPs through reserved IP addresses.

**Reserve or Identify Existing IP:**

```bash
# List existing static IPs
gcloud compute addresses list

# Or create a new static IP
gcloud compute addresses create traefik-ip --region us-central1
```

Update `traefik-values.yaml`:

```yaml
service:
  type: LoadBalancer
  spec:
    loadBalancerIP: "<your-static-ip>"
```

### Other Cloud Providers

For other cloud providers, check their Kubernetes LoadBalancer documentation:

- **DigitalOcean:** Supports `loadBalancerIP` with floating IPs
- **Linode:** Supports `loadBalancerIP` specification
- **Bare Metal:** Use MetalLB with IP address pools

---

## Step 3: Install Traefik

With your values file configured, install Traefik:

```bash
# Install Traefik in a dedicated namespace
kubectl create namespace traefik

helm install traefik traefik/traefik \
  --namespace traefik \
  --values traefik-values.yaml
```

### Verify Installation

```bash
# Check Traefik pods are running
kubectl get pods -n traefik

# Check the service has the correct LoadBalancer IP
kubectl get svc -n traefik traefik

# Check Traefik logs
kubectl logs -n traefik deployment/traefik
```

Expected output:

```bash
NAME      TYPE           CLUSTER-IP      EXTERNAL-IP      PORT(S)
traefik   LoadBalancer   10.0.123.45     <your-lb-ip>     80:<nodeport>/TCP,443:<nodeport>/TCP
```

---

## Step 4: Test with Sample Application

Before migrating production workloads, test Traefik with a sample application using the `traefik/whoami` image.

### Deploy Whoami Application

```yaml tab="whoami-deployment.yaml"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: whoami
  namespace: default
  labels:
    app: whoami
spec:
  replicas: 2
  selector:
    matchLabels:
      app: whoami
  template:
    metadata:
      labels:
        app: whoami
    spec:
      containers:
        - name: whoami
          image: traefik/whoami
          ports:
            - containerPort: 80
              name: http

---
apiVersion: v1
kind: Service
metadata:
  name: whoami
  namespace: default
spec:
  selector:
    app: whoami
  ports:
    - name: http
      port: 80
      targetPort: http
```

Apply the deployment:

```bash
kubectl apply -f whoami-deployment.yaml
```

### Create Test Ingress with NGINX Annotations

```yaml tab="whoami-ingress.yaml"
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: whoami-test
  namespace: default
  annotations:
    # NGINX annotation for SSL redirect (supported by Traefik)
    nginx.ingress.kubernetes.io/ssl-redirect: "false"
    # NGINX annotation for CORS (supported by Traefik)
    nginx.ingress.kubernetes.io/enable-cors: "true"
spec:
  ingressClassName: nginx
  rules:
    - host: whoami.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: whoami
                port:
                  number: 80
```

Apply the Ingress:

```bash
kubectl apply -f whoami-ingress.yaml
```

### Validate Traffic Routing

Test that Traefik correctly routes traffic to the whoami service:

```bash
# Test using the LoadBalancer IP
curl -H "Host: whoami.example.com" http://<traefik-lb-ip>/

# Or add an entry to /etc/hosts for testing
echo "<traefik-lb-ip> whoami.example.com" | sudo tee -a /etc/hosts
curl http://whoami.example.com/
```

Expected output:

```bash
Hostname: whoami-xxxxxxxxx-xxxxx
IP: 10.244.x.x
RemoteAddr: 10.244.y.y:xxxxx
GET / HTTP/1.1
Host: whoami.example.com
User-Agent: curl/7.x.x
Accept: */*
```

If `publishService` is enabled in your values file, `kubectl get ingress whoami-test` should now show the Traefik LoadBalancer IP in the `ADDRESS` column.

!!! success "Test Successful"

    If you see the whoami response, Traefik is correctly routing traffic using NGINX annotations. You're ready to migrate production Ingresses.

---

## Step 5: Migrate Ingress Resources

Now that Traefik is validated, migrate your production Ingress resources from NGINX to Traefik.

### Migration Approaches

**Option A: In-Place Update (Quick Switchover)**

Traefik's Ingress NGINX provider already watches `ingressClassName: nginx`, so you usually do **not** need to change Ingress manifests. Ensure each Ingress explicitly sets that class (or the legacy `kubernetes.io/ingress.class: nginx` annotation) so Traefik picks it up once the NGINX service is removed:

```bash
# Update a single Ingress
kubectl patch ingress <ingress-name> -n <namespace> -p '{"spec":{"ingressClassName":"nginx"}}'

# Or edit manually
kubectl edit ingress <ingress-name> -n <namespace>
```

**Option B: Gradual Migration (Lower Risk)**

Create duplicate Ingresses with different hostnames for testing, then switch DNS:

1. Create new Ingress resources with `-traefik` suffix
2. Test using the new hostname
3. Update DNS records
4. Delete old NGINX Ingresses

### Common NGINX Annotations Supported by Traefik

Traefik supports many commonly used NGINX annotations, including authentication, TLS/SSL, session affinity, CORS, and routing configurations. Some annotations have partial support or behavioral differences compared to NGINX.

For a complete list of supported annotations, limitations, and workarounds, see the [Ingress NGINX Routing Configuration](../reference/routing-configuration/kubernetes/ingress-nginx.md) documentation.

### Example: Migrate Ingress with Common Annotations

```yaml tab="Before (NGINX)"
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: myapp
  namespace: production
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/enable-cors: "true"
    nginx.ingress.kubernetes.io/cors-allow-origin: "https://example.com"
    nginx.ingress.kubernetes.io/affinity: "cookie"
    nginx.ingress.kubernetes.io/session-cookie-name: "route"
spec:
  ingressClassName: nginx  # ← NGINX IngressClass
  rules:
    - host: myapp.example.com
      http:
        paths:
          - path: /api
            pathType: Prefix
            backend:
              service:
                name: myapp-service
                port:
                  number: 8080
```

```yaml tab="After (Traefik)"

apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: myapp
  namespace: production
  annotations:
    # Same NGINX annotations work with Traefik
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/enable-cors: "true"
    nginx.ingress.kubernetes.io/cors-allow-origin: "https://example.com"
    nginx.ingress.kubernetes.io/affinity: "cookie"
    nginx.ingress.kubernetes.io/session-cookie-name: "route"
spec:
  ingressClassName: nginx  # ← Same class, Traefik watches for it
  rules:
    - host: myapp.example.com
      http:
        paths:
          - path: /api
            pathType: Prefix
            backend:
              service:
                name: myapp-service
                port:
                  number: 8080
```

The same Ingress configuration works with Traefik, you only need to ensure `ingressClassName` is set correctly.

!!! tip "No Changes Required"

    In most cases, you don't need to modify your Ingress resources at all. Traefik automatically translates NGINX annotations into its dynamic configuration.

!!! warning "NGINX Admission Webhook Cleanup"

    After you stop using the NGINX controller, delete the `ingress-nginx-admission` validating/mutating webhooks (and the admission service) or future Ingress edits can fail with webhook timeouts.

### Existing TLS Secrets

Existing TLS configurations continue to work with Traefik:

- Keep `spec.tls` entries exactly as-is; Traefik terminates TLS using the referenced secrets.
- TLS secrets must stay in the same namespace as the Ingress and list the hostnames you serve.
- NGINX `ssl-redirect` / `force-ssl-redirect` annotations are honored (unless globally forced).
- After switching, verify with `kubectl describe ingress <name>` and `curl -k https://<host>/`.

---

## Step 6: Monitor and Validate

After migrating Ingress resources, monitor traffic to ensure everything is working correctly.

### Validation Commands

```bash
# Check Ingress resources are discovered by Traefik
kubectl get ingress --all-namespaces

# View Traefik logs for routing decisions
kubectl logs -n traefik deployment/traefik -f

# Test application endpoints
curl -v https://myapp.example.com/api/health
```

### Common Issues and Troubleshooting

**Issue: Ingresses not discovered by Traefik**

Check IngressClass configuration:

```bash
# Verify IngressClass exists
kubectl get ingressclass

# Check Traefik is watching the correct IngressClass
kubectl logs -n traefik deployment/traefik | grep -i "ingress"
```

**Issue: Annotation not working as expected**

Some NGINX annotations have behavioral differences. Check the [limitations documentation](../reference/routing-configuration/kubernetes/ingress-nginx.md#limitations) for details.

**Issue: TLS certificates not working**

Ensure TLS secrets are in the same namespace as the Ingress:

```bash
# List secrets in the namespace
kubectl get secrets -n <namespace>

# Check TLS secret format
kubectl get secret <tls-secret-name> -n <namespace> -o yaml
```

---

## Step 7: Decommission NGINX Ingress Controller

Once all traffic has been successfully migrated and validated, you can safely remove the NGINX Ingress Controller.

### Verification Checklist

Before decommissioning NGINX, ensure:

- All Ingresses are routing through Traefik
- No errors in application logs
- Monitoring shows healthy traffic patterns
- TLS certificates are working correctly
- Authentication and authorization are functioning
- NGINX admission webhook is no longer needed (Traefik is the only Ingress controller)

### Remove NGINX Resources

!!! info
    
    In this section, we assume that the NGINX Ingress Controller was installed in the namespace `ingress-nginx`. If it was installed in a different namespace, you need to replace `ingress-nginx` with the namespace name.
  
If NGINX Ingress Controller was installed via Helm, run the following command to uninstall the chart and clean up chart-managed resources:

```bash
helm uninstall ingress-nginx -n ingress-nginx
```

!!! warning "IngressClass Deleted by Helm"

    When you run `helm uninstall`, the `nginx` IngressClass is also deleted. If you followed the Helm values configuration in [Step 1](#step-1-install-traefik-with-helm) (which includes `extraObjects` to create the IngressClass), Traefik will continue to work. If you did not include the `extraObjects` section, you must manually create the IngressClass:

    ```bash
    kubectl apply -f - <<EOF
    apiVersion: networking.k8s.io/v1
    kind: IngressClass
    metadata:
      name: nginx
    spec:
      controller: k8s.io/ingress-nginx
    EOF
    ```

If NGINX Ingress Controller was installed manually or need to hurry Load Balancer IP release, run the explicit deletes:

```bash
# Scale down NGINX deployment
kubectl scale deployment -n ingress-nginx ingress-nginx-controller --replicas=0

# Wait and monitor for issues (recommended: 24-48 hours)

# Delete NGINX deployment
kubectl delete deployment -n ingress-nginx ingress-nginx-controller

# Delete NGINX service
kubectl delete svc -n ingress-nginx ingress-nginx-controller
kubectl delete svc -n ingress-nginx ingress-nginx-controller-admission

# Delete NGINX ConfigMaps
kubectl delete configmap -n ingress-nginx ingress-nginx-controller

# Delete NGINX RBAC resources
kubectl delete clusterrole ingress-nginx
kubectl delete clusterrolebinding ingress-nginx
kubectl delete role -n ingress-nginx ingress-nginx
kubectl delete rolebinding -n ingress-nginx ingress-nginx
kubectl delete serviceaccount -n ingress-nginx ingress-nginx

# Delete NGINX admission webhooks (prevents future Ingress edits from failing)
kubectl delete validatingwebhookconfiguration ingress-nginx-admission
kubectl delete mutatingwebhookconfiguration ingress-nginx-admission --ignore-not-found

# Delete the namespace (if dedicated to NGINX)
kubectl delete namespace ingress-nginx
```

!!! note "LoadBalancer IP release timing"

    Cloud providers may hold on to the NGINX LoadBalancer IP for a short time after deleting the Service. Deleting the NGINX service early (or letting Helm uninstall it) speeds up reusing that IP for Traefik, but actual release timing depends on the provider.

!!! success "Migration Complete"

    Congratulations! You have successfully migrated from NGINX Ingress Controller to Traefik.

---

## Next Steps

Now that you've successfully migrated to Traefik, explore additional capabilities:

**Learn More About Traefik:**

- [Kubernetes Ingress NGINX Install Configuration](../reference/install-configuration/providers/kubernetes/kubernetes-ingress-nginx.md) - Detailed provider configuration
- [Kubernetes Ingress NGINX Routing Configuration](../reference/routing-configuration/kubernetes/ingress-nginx.md) - Routing rules and annotation support
- [HTTP Middlewares](../reference/routing-configuration/http/middlewares/overview.md) - Extend functionality beyond NGINX annotations
- [TLS Configuration](../reference/routing-configuration/http/tls/overview.md) - Advanced TLS and certificate management

**Enhance Your Setup:**

- Enable [metrics](../reference/install-configuration/observability/metrics.md) and [tracing](../reference/install-configuration/observability/tracing.md)
- Configure [access logs](../reference/install-configuration/observability/logs-and-accesslogs.md) for observability
- Explore [Traefik Middlewares](../reference/routing-configuration/http/middlewares/overview.md) for advanced traffic management
- Consider [Traefik Hub](https://traefik.io/traefik-hub/) for enterprise features like AI & API Gateway, API Management, and advanced security

---

## Feedback and Support

If you encounter issues during migration or have suggestions for improving this guide:

- **Report Issues:** [GitHub Issues](https://github.com/traefik/traefik/issues)
- **Community Support:** [Traefik Community Forum](https://community.traefik.io/)
- **Enterprise Support:** [Traefik Labs Commercial Support](https://traefik.io/pricing/)

We welcome contributions to improve this migration guide. See our [contribution guidelines](../contributing/submitting-pull-requests.md) to get started.
