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

- Prepare the Traefik Helm values configuration
- Delete the NGINX LoadBalancer service to release the IP (downtime starts)
- Install Traefik to claim the released IP (downtime ends)
- Verify traffic routing through Traefik
- Uninstall NGINX Ingress Controller and recreate the IngressClass

**Expected Downtime:** Brief downtime (typically seconds to a few minutes) during the switchover, depending on cloud provider IP release timing. If using DNS-based migration instead of IP retention, downtime depends on DNS TTL.

---

## Step 1: Prepare Traefik Configuration

Before performing the switchover, prepare your Traefik Helm configuration. Traefik is installed using Helm, which automatically creates all necessary RBAC resources and configures the Kubernetes Ingress NGINX provider.

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

# Service configuration (see Load Balancer IP Retention section below)
service:
  enabled: true
  type: LoadBalancer
  # Additional cloud-provider specific configuration
  annotations: {}
  spec: {}
    # Configure loadBalancerIP here (see Load Balancer IP Retention section)
    # loadBalancerIP: "<your-existing-ip>"
```

!!! info "RBAC Included"

    The Traefik Helm chart automatically creates all necessary RBAC resources:

    - ServiceAccount
    - ClusterRole with permissions for Ingresses, Services, Endpoints, Secrets
    - ClusterRoleBinding

    No manual RBAC configuration is required.

!!! note "IngressClass"

    At this point, the `nginx` IngressClass already exists (created by the NGINX Ingress Controller). Traefik will use this existing IngressClass to discover your Ingresses. After uninstalling NGINX in [Step 4](#step-4-uninstall-nginx-and-recreate-ingressclass), you will need to recreate this IngressClass.

### Load Balancer IP Retention

To minimize downtime during migration, retain your existing Load Balancer IP address. This prevents DNS propagation delays and ensures continuous service availability.

#### General Approach

The `loadBalancerIP` field allows you to specify a static IP for the LoadBalancer service. However, cloud providers handle this differently.

!!! warning "Cloud Provider Variations"

    Load balancer IP retention strategies vary by cloud provider. Follow the appropriate section for your environment.

!!! warning "ServiceLB and hostPort-based environments"

    On k3s/k3d (and any environment that implements `Service.Type=LoadBalancer` by binding host ports such as Klipper/ServiceLB), only one LoadBalancer service can listen on ports 80/443 at a time. Delete the NGINX LoadBalancer service before creating Traefik's service, or use different ports, otherwise Traefik's service will stay in `Pending`.

#### AWS Elastic Load Balancer

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
4. [Uninstall NGINX](#step-4-uninstall-nginx-and-recreate-ingressclass)

#### Azure Load Balancer

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

#### GCP Load Balancer

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

#### Other Cloud Providers

For other cloud providers, check their Kubernetes LoadBalancer documentation:

- **DigitalOcean:** Supports `loadBalancerIP` with floating IPs
- **Linode:** Supports `loadBalancerIP` specification
- **Bare Metal:** Use MetalLB with IP address pools

---

## Step 2: Switch Traffic to Traefik

This step performs the actual switchover. There will be brief downtime while the LoadBalancer IP transitions from NGINX to Traefik.

### Delete NGINX LoadBalancer Service

First, delete the NGINX service to release the LoadBalancer IP:

```bash
# Delete NGINX service to release the IP (downtime starts)
kubectl delete svc -n ingress-nginx ingress-nginx-controller
```

### Install Traefik

Immediately install Traefik to claim the released IP:

```bash
# Create Traefik namespace
kubectl create namespace traefik

# Install Traefik (downtime ends when LoadBalancer IP is assigned)
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

# Check Traefik logs for any errors
kubectl logs -n traefik deployment/traefik
```

Expected output:

```bash
NAME      TYPE           CLUSTER-IP      EXTERNAL-IP      PORT(S)
traefik   LoadBalancer   10.0.123.45     <your-lb-ip>     80:<nodeport>/TCP,443:<nodeport>/TCP
```

---

## Step 3: Verify Traffic Routing

After Traefik is running, verify that your existing Ingresses are being served correctly.

### Test Existing Applications

Test your production endpoints to confirm traffic is flowing through Traefik:

```bash
# Test using the LoadBalancer IP with a custom Host header
curl --resolve myapp.example.com:80:<traefik-lb-ip> http://myapp.example.com/

# Test HTTPS
curl --resolve myapp.example.com:443:<traefik-lb-ip> -k https://myapp.example.com/
```

### Verify Ingress Status

If `publishService` is enabled in your values file, check that Ingress resources show the Traefik LoadBalancer IP:

```bash
kubectl get ingress --all-namespaces
```

The `ADDRESS` column should display the Traefik LoadBalancer IP.

### Check Traefik Logs

Monitor Traefik logs for routing decisions and any errors:

```bash
kubectl logs -n traefik deployment/traefik -f
```

---

### Common NGINX Annotations Supported by Traefik

Traefik supports many commonly used NGINX annotations, including authentication, TLS/SSL, session affinity, CORS, and routing configurations. Some annotations have partial support or behavioral differences compared to NGINX.

#### Ingress Configuration (No Changes Required)

The key benefit of the Traefik Kubernetes Ingress NGINX provider is that your existing Ingress resources work without modification:

```yaml tab="Your Existing Ingress (No Changes Needed)"
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: myapp
  namespace: production
  annotations:
    # These NGINX annotations are automatically translated by Traefik
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/enable-cors: "true"
    nginx.ingress.kubernetes.io/cors-allow-origin: "https://example.com"
    nginx.ingress.kubernetes.io/affinity: "cookie"
    nginx.ingress.kubernetes.io/session-cookie-name: "route"
spec:
  ingressClassName: nginx  # ← Traefik watches this class
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

#### Supported Annotations

For a complete list of supported annotations, limitations, and behavioral differences, see the [Ingress NGINX Routing Configuration](../reference/routing-configuration/kubernetes/ingress-nginx.md) documentation.

#### Existing TLS Secrets

Existing TLS configurations continue to work with Traefik:

- Keep `spec.tls` entries exactly as-is; Traefik terminates TLS using the referenced secrets
- TLS secrets must stay in the same namespace as the Ingress
- NGINX `ssl-redirect` / `force-ssl-redirect` annotations are honored
- Verify with `kubectl describe ingress <name>` and `curl -k https://<host>/`

---

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

## Step 4: Uninstall NGINX and Recreate IngressClass

Once traffic is verified flowing through Traefik, remove the NGINX Ingress Controller and recreate the IngressClass.

!!! info "Namespace Assumption"
    
    This section assumes NGINX Ingress Controller was installed in the namespace `ingress-nginx`. Replace with your actual namespace if different.

### Uninstall NGINX Ingress Controller

If NGINX was installed via Helm:

```bash
helm uninstall ingress-nginx -n ingress-nginx
```

If NGINX was installed manually, delete resources explicitly:

```bash
# Delete NGINX deployment
kubectl delete deployment -n ingress-nginx ingress-nginx-controller

# Delete remaining NGINX services
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

### Recreate the IngressClass

When NGINX is uninstalled, the `nginx` IngressClass is also deleted. Traefik needs this IngressClass to continue recognizing your existing Ingresses. Recreate it:

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

Verify the IngressClass exists:

```bash
kubectl get ingressclass
```

!!! success "Migration Complete"

    Congratulations! You have successfully migrated from NGINX Ingress Controller to Traefik. Your existing Ingresses with `ingressClassName: nginx` continue to work, now served by Traefik.

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
