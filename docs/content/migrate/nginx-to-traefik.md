---
title: "Migrate from Ingress NGINX Controller to Traefik"
description: "Step-by-step guide to migrate from Kubernetes Ingress NGINX Controller to Traefik with zero downtime and annotation compatibility."
---

# Migrate from Ingress NGINX Controller to Traefik

How to migrate from Ingress NGINX Controller to Traefik with zero downtime.
{: .subtitle }

---

!!! danger "Ingress NGINX Controller Retirement"

    The Kubernetes Ingress NGINX Controller project has announced its retirement in **March 2026**. After this date:

    - No new releases or updates
    - No security patches
    - No bug fixes

    For more information, see the [official Kubernetes blog announcement](https://kubernetes.io/blog/2025/11/11/ingress-nginx-retirement).

## What You Will Achieve

By completing this migration, your existing Ingress resources will work with Traefik without any modifications. The Traefik Kubernetes Ingress NGINX Provider automatically translates NGINX annotations into Traefik configuration:

```yaml tab="Your Existing Ingress (No Changes Needed)"
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: myapp
  annotations:
    # These NGINX annotations are automatically translated by Traefik
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
    nginx.ingress.kubernetes.io/enable-cors: "true"
    nginx.ingress.kubernetes.io/cors-allow-origin: "https://example.com"
    nginx.ingress.kubernetes.io/affinity: "cookie"
    nginx.ingress.kubernetes.io/session-cookie-name: "route"
spec:
  ingressClassName: nginx  # ← Traefik will watch this class
  rules:
    - host: myapp.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: whoami
                port:
                  number: 80

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: whoami
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

---
apiVersion: v1
kind: Service
metadata:
  name: whoami
spec:
  selector:
    app: whoami
  ports:
    - protocol: TCP
      port: 80
      targetPort: 80
```

For a complete list of supported annotations and behavioral differences, see the [Ingress NGINX Routing Configuration](../reference/routing-configuration/kubernetes/ingress-nginx.md) documentation.

!!! info "Traefik Version Requirement"

    The Kubernetes Ingress NGINX provider requires **Traefik v3.6.2 or later**.

---

## Prerequisites

Before starting the migration, ensure you have:

- **Existing Ingress NGINX Controller** running in your Kubernetes cluster
- **Kubernetes cluster access** with `kubectl` configured 
- **Cluster support for running multiple LoadBalancer services** on ports 80/443 simultaneously
- **Helm**
- **Cluster admin permissions** to create RBAC resources
- **Backup of critical configurations** (Ingress resources, ConfigMaps, Secrets)

!!! tip "Backup Recommendations"

    ```bash
    # Export all Ingress resources
    kubectl get ingress --all-namespaces -o yaml > ingress-backup.yaml

    # Export NGINX ConfigMaps
    kubectl get configmap --all-namespaces -l app.kubernetes.io/name=ingress-nginx -o yaml > nginx-configmaps.yaml
    ```

---

## Migration Strategy Overview

This migration achieves **zero downtime** by running Traefik alongside NGINX. Both controllers serve the same Ingress resources simultaneously, allowing you to progressively shift traffic before removing NGINX.

```text
Current:     DNS → LoadBalancer → NGINX → Your Services

Migration:   DNS → LoadBalancer → NGINX  → Your Services
                 → LoadBalancer → Traefik → Your Services

Final:       DNS → LoadBalancer → Traefik → Your Services
```

**Migration Flow:**

1. Install Traefik alongside NGINX (both serving traffic in parallel)
2. Add Traefik LoadBalancer to DNS (if you choose DNS option; cf. step 3)
3. Progressively shift traffic from NGINX to Traefik
4. Remove NGINX from DNS, preserve the IngressClass, and uninstall

---

## Step 1: Install Traefik Alongside NGINX

??? info "Install Ingress NGINX Controller"

    If you have not installed Ingress NGINX Controller yet, you can set up a fresh Ingress NGINX Controller installation following the instructions below:

    ### Install Ingress NGINX Controller

    ```bash
    helm upgrade --install ingress-nginx ingress-nginx \
      --repo https://kubernetes.github.io/ingress-nginx \
      --namespace ingress-nginx --create-namespace
    ```
Install Traefik with the Kubernetes Ingress NGINX provider enabled. Both controllers will serve the same Ingress resources simultaneously.

### Add Traefik Helm Repository

```bash
helm repo add traefik https://traefik.github.io/charts
helm repo update
```

### Install Traefik

```bash
helm upgrade --install traefik traefik/traefik \
  --namespace traefik --create-namespace \
  --set providers.kubernetesIngressNginx.enabled=true
```

Or using a [values file](https://github.com/traefik/traefik-helm-chart/blob/master/traefik/VALUES.md) for more configuration:

```yaml tab="traefik-values.yaml"
...
providers:
  kubernetesIngressNginx:
    enabled: true
 ...
```

```bash
helm upgrade --install traefik traefik/traefik \
  --namespace traefik --create-namespace \
  --values traefik-values.yaml
```

### Verify Both Controllers Are Running

```bash
# Check NGINX pods
kubectl get pods -n ingress-nginx

# Check Traefik pods
kubectl get pods -n traefik

# Check both services have LoadBalancer IPs
kubectl get svc -n ingress-nginx ingress-nginx-controller
kubectl get svc -n traefik traefik
```

At this point, both NGINX and Traefik are running and can serve the same Ingress resources. Traffic is still flowing only through NGINX since DNS points to the NGINX LoadBalancer.

---

## Step 2: Verify Traefik Is Handling Traffic

Before adding Traefik to DNS, verify it correctly serves your Ingress resources.

### Test via Traefik's LoadBalancer IP

Get Traefik's LoadBalancer IP and use `--resolve` to test without changing DNS:

```bash
# Get LoadBalancer IPs
NGINX_IP=$(kubectl get svc -n ingress-nginx ingress-nginx-controller -o go-template='{{ $ing := index .status.loadBalancer.ingress 0 }}{{ if $ing.ip }}{{ $ing.ip }}{{ else }}{{ $ing.hostname }}{{ end }}')
TRAEFIK_IP=$(kubectl get svc -n traefik traefik -o go-template='{{ $ing := index .status.loadBalancer.ingress 0 }}{{ if $ing.ip }}{{ $ing.ip }}{{ else }}{{ $ing.hostname }}{{ end }}')
echo -e "Nginx IP: $NGINX_IP\nTraefik IP: $TRAEFIK_IP"

# Test HTTP for both
FQDN=myapp.example.com
# Observe HTTPS redirections:
curl --connect-to "${FQDN}:80:${NGINX_IP}:80" "http://${FQDN}" -D -
curl --connect-to "${FQDN}:80:${TRAEFIK_IP}:80" "http://${FQDN}" -D - # note X-Forwarded-Server which should be traefik

# Test HTTPS
curl --connect-to "${FQDN}:443:${NGINX_IP}:443" "https://${FQDN}"
curl --connect-to "${FQDN}:443:${TRAEFIK_IP}:443" "https://${FQDN}"
```

!!! warning "TLS Certificates During Migration"

    Both NGINX and Traefik must serve valid TLS certificates for HTTPS tests to succeed. Since Traefik is not publicly exposed during this verification phase, **Let's Encrypt HTTP challenge will not work**.

    Your options for TLS certificates during migration:

    - **Existing certificates via `tls.secretName`** - If you use cert-manager or another external tool, your existing TLS secrets referenced in `spec.tls` will work with both controllers
    - **Let's Encrypt DNS challenge** - Configure Traefik's [ACME DNS challenge](../reference/install-configuration/tls/certificate-resolvers/acme.md#dnschallenge) to obtain certificates without public exposure

    Avoid using `curl -k` (skip certificate verification) as this masks TLS configuration issues that could cause problems after migration.

### Verify Ingress Discovery

Check Traefik logs to confirm it discovered your Ingress resources:

```bash
kubectl logs -n traefik deployment/traefik | grep -i "ingress"
```

---

## Step 3: Shift Traffic to Traefik

With both controllers running and verified, progressively shift traffic from NGINX to Traefik.

### Option A: DNS-Based Migration

Add the Traefik LoadBalancer IP to your DNS records alongside NGINX. This allows both controllers to receive traffic.

**Get LoadBalancer addresses:**

```bash
# NGINX LoadBalancer
echo $(kubectl get svc -n ingress-nginx ingress-nginx-controller -o go-template='{{ $ing := index .status.loadBalancer.ingress 0 }}{{ if $ing.ip }}{{ $ing.ip }}{{ else }}{{ $ing.hostname }}{{ end }}')

# Traefik LoadBalancer
echo $(kubectl get svc -n traefik traefik -o go-template='{{ $ing := index .status.loadBalancer.ingress 0 }}{{ if $ing.ip }}{{ $ing.ip }}{{ else }}{{ $ing.hostname }}{{ end }}')
```

**Progressive DNS migration:**

1. **Add Traefik to DNS** - Add the Traefik LoadBalancer IP to your DNS records (both IPs now receive traffic via round-robin)
2. **Monitor** - Observe traffic patterns on both controllers
3. **Remove NGINX from DNS** - Once confident, remove the NGINX LoadBalancer IP from DNS
4. **Wait for DNS propagation** - Allow time for DNS caches to expire
5. **Uninstall NGINX** - Proceed to [Step 4](#step-4-uninstall-ingress-nginx-controller)

!!! warning "DNS TTL May Not Be Respected"

    Some ISPs ignore DNS TTL values to reduce traffic costs, caching records longer than specified. After removing NGINX from DNS, keep NGINX running for at least 24-48 hours before uninstalling to avoid dropping traffic from users whose ISPs have stale DNS caches.

??? info "ExternalDNS Users"

    If you use [ExternalDNS](https://github.com/kubernetes-sigs/external-dns) to automatically manage DNS records based on Ingress status, both NGINX and Traefik will compete to update the Ingress status with their LoadBalancer IPs when `publishService` is enabled. Traefik typically wins because it updates faster, which can cause unexpected traffic shifts.

    **Recommended approach for ExternalDNS:**

    1. **[Install Traefik](#step-1-install-traefik-alongside-nginx) with `publishService` disabled**:

        ```yaml
        # traefik-values.yaml
        providers:
          kubernetesIngressNginx:
            enabled: true
            publishService:
              enabled: false  # Disable to prevent status updates
        ```

    2. **Test Traefik** using [port-forward](#step-2-verify-traefik-is-handling-traffic) or a separate test hostname

    3. **Switch DNS via NGINX** - Configure NGINX to publish Traefik's service address:

        ```yaml
        # nginx-values.yaml
        controller:
          publishService:
            pathOverride: "traefik/traefik"  # Points to Traefik's service
        ```

        This makes NGINX update the Ingress status with Traefik's LoadBalancer IP, causing ExternalDNS to point traffic to Traefik.

    4. **Verify traffic flows through Traefik** - At this point, you can still rollback by removing the `pathOverride`

    5. **[Enable `publishService` on Traefik](#step-1-install-traefik-alongside-nginx)** and [uninstall NGINX](#step-5-uninstall-nginx-ingress-controller)

### Option B: External Load Balancer with Weighted Traffic

For more control over traffic distribution, use an external load balancer (like Traefik, Cloudflare, AWS ALB, or a dedicated load balancer) in front of both Kubernetes LoadBalancers.

!!! note "Infrastructure Prerequisite"

    This option assumes you already have an external load balancer in your infrastructure, or are willing to set one up **before** starting the migration. Adding an external load balancer is a significant infrastructure change that should be planned and tested separately from the ingress controller migration.

**Setup:**

1. Create an external load balancer pointing to the NGINX Kubernetes LoadBalancer
2. Update DNS to point to the external load balancer
3. Add the Traefik Kubernetes LoadBalancer to the external load balancer with a low weight (e.g., 10%)
4. Gradually increase Traefik's weight while decreasing NGINX's weight
5. Once NGINX receives no traffic, uninstall it

**Example weight progression:**

| Phase | NGINX Weight | Traefik Weight | Duration |
|-------|-------------|----------------|----------|
| Initial | 100% | 0% | - |
| Start | 90% | 10% | 1 hour |
| Increase | 50% | 50% | 2 hour |
| Near-complete | 10% | 90% | 4 hour |
| Final | 0% | 100% | - |

!!! tip "External Load Balancer Options"

    - **Cloudflare Load Balancing** - Traffic steering with health checks
    - **AWS Global Accelerator** - Weighted routing across endpoints
    - **Google Cloud Load Balancing** - Traffic splitting
    - **Traefik / HAProxy / NGINX (external)** - Self-hosted option with weighted backends
    - ...

### LoadBalancer IP Retention

If you want Traefik to eventually use the same LoadBalancer IP as NGINX (to simplify DNS management), you can transfer the IP after the migration. Since Traefik is already running with its own LoadBalancer, this can be done with zero downtime.

**Zero-downtime IP transfer process:**

1. Traefik is already running with its own LoadBalancer IP (from Step 1)
2. Add Traefik's LoadBalancer IP to DNS (traffic now goes to both NGINX and Traefik)
3. Remove NGINX's IP from DNS and wait for propagation
4. Delete NGINX's LoadBalancer service to release the IP
5. Upgrade Traefik to claim the released IP
6. (Optional) Remove Traefik's old IP from DNS once the new IP is active

This way, traffic is always flowing to Traefik during the IP transfer.

**Get your current NGINX LoadBalancer IP:**

```bash
kubectl get svc -n ingress-nginx ingress-nginx-controller -o go-template='{{ $ing := index .status.loadBalancer.ingress 0 }}{{ if $ing.ip }}{{ $ing.ip }}{{ else }}{{ $ing.hostname }}{{ end }}'
```

??? note "AWS (Network Load Balancer with Elastic IPs)"

    AWS does not support static IPs for Classic Load Balancers. Use Network Load Balancers (NLB) with Elastic IPs instead. This requires the [AWS Load Balancer Controller](https://kubernetes-sigs.github.io/aws-load-balancer-controller/) to be installed in your cluster.

    **Pre-allocate Elastic IPs (one per availability zone):**

    ```bash
    aws ec2 allocate-address --domain vpc --region <your-region>
    # Note the AllocationId (eipalloc-xxx) for each EIP
    ```

    **Update `traefik-values.yaml`:**

    ```yaml
    service:
      type: LoadBalancer
      loadBalancerClass: service.k8s.aws/nlb  # Requires AWS Load Balancer Controller
      annotations:
        service.beta.kubernetes.io/aws-load-balancer-type: "external"
        service.beta.kubernetes.io/aws-load-balancer-nlb-target-type: "ip"
        service.beta.kubernetes.io/aws-load-balancer-eip-allocations: "eipalloc-xxx,eipalloc-yyy"
    ```

    For more details, see the [AWS Load Balancer Controller annotations documentation](https://kubernetes-sigs.github.io/aws-load-balancer-controller/latest/guide/service/annotations/).

??? note "Azure"

    Azure supports static public IPs for Load Balancers.

    **Identify existing public IP:**

    ```bash
    az network public-ip list --resource-group <your-resource-group> \
      --query "[?ipAddress=='<your-ip>'].name" -o tsv
    ```

    **Update `traefik-values.yaml`:**

    ```yaml
    service:
      type: LoadBalancer
      annotations:
        # Only needed if the public IP is in a different resource group than the AKS cluster
        service.beta.kubernetes.io/azure-load-balancer-resource-group: "<public-ip-resource-group>"
      spec:
        loadBalancerIP: "<your-existing-ip>"
    ```

    For more details, see the [Azure AKS static IP documentation](https://learn.microsoft.com/en-us/azure/aks/static-ip).

??? note "GCP"

    GCP supports static IPs through reserved regional IP addresses.

    **Reserve or identify existing IP:**

    ```bash
    # List existing static IPs
    gcloud compute addresses list

    # Or reserve a new regional static IP (must be in the same region as your GKE cluster)
    gcloud compute addresses create traefik-ip --region <your-cluster-region>
    ```

    **Update `traefik-values.yaml`:**

    ```yaml
    service:
      type: LoadBalancer
      spec:
        loadBalancerIP: "<your-static-ip>"
    ```

    For more details, see the [GKE LoadBalancer Service parameters documentation](https://cloud.google.com/kubernetes-engine/docs/concepts/service-load-balancer-parameters).

??? note "Other Cloud Providers"

    - **DigitalOcean:** Supports `loadBalancerIP` with floating IPs
    - **Linode:** Supports `loadBalancerIP` specification
    - **Bare Metal (MetalLB):** Use IP address pools

**Transfer the IP:**

Once DNS is pointing to Traefik and your values are configured with the target IP:

```bash
# Ensure Traefik is already receiving traffic via its current LoadBalancer
kubectl get svc -n traefik traefik

# Delete NGINX LoadBalancer service to release the IP
kubectl delete svc -n ingress-nginx ingress-nginx-controller

# Upgrade Traefik to claim the released IP
helm upgrade traefik traefik/traefik \
  --namespace traefik \
  --values traefik-values.yaml

# Verify Traefik now has the old NGINX IP
kubectl get svc -n traefik traefik
```

!!! tip "Zero Downtime During Helm Upgrade"

    The Helm upgrade only restarts the Traefik pod, not the LoadBalancer service. Traefik uses a `RollingUpdate` deployment strategy by default, so the new pod starts before the old one terminates. For additional safety, configure high availability:

    ```yaml
    # In traefik-values.yaml
    deployment:
      replicas: 2

    # Spread pods across nodes to survive node failures
    affinity:
      podAntiAffinity:
        requiredDuringSchedulingIgnoredDuringExecution:
          - labelSelector:
              matchLabels:
                app.kubernetes.io/name: traefik
                app.kubernetes.io/instance: traefik
            topologyKey: kubernetes.io/hostname

    # Ensure at least one pod is always available during disruptions
    podDisruptionBudget:
      enabled: true
      minAvailable: 1
    ```

    With multiple replicas spread across nodes and a PodDisruptionBudget, at least one pod is always running during upgrades and node maintenance.

---

## Step 4: Uninstall Ingress NGINX Controller

Once NGINX is no longer receiving traffic, remove it from your cluster. Before uninstalling, you must ensure the `nginx` IngressClass is preserved. Traefik needs it to continue discovering your Ingresses.

### Preserve the IngressClass

??? note "If NGINX Was Installed via Helm"

    Add the `helm.sh/resource-policy: keep` annotation to tell Helm to preserve the IngressClass:

    ```bash
    # Add the required annotation
    helm upgrade ingress-nginx ingress-nginx \
      --repo https://kubernetes.github.io/ingress-nginx \
      --namespace ingress-nginx \
      --reuse-values \
      --set-json 'controller.ingressClassResource.annotations={"helm.sh/resource-policy": "keep"}'
    # Check that the annotation is really here
    kubectl describe ingressclass nginx
    ```

    The `--reuse-values` flag is critical - it preserves all your existing NGINX configuration. Without it, Helm would reset everything to defaults, potentially breaking your setup.

    !!! info "kubectl annotate/patch/edit does not work"

        Adding the annotation via `kubectl annotate`, `kubectl patch`, or `kubectl edit` will not preserve the IngressClass. Helm stores its release state internally and checks annotations from its internal manifest, not the live cluster state. Only `helm upgrade` updates Helm's internal state.

??? note "If NGINX Was Installed via GitOps (ArgoCD, Flux)"

    Ensure the `nginx` IngressClass is defined as a standalone resource in your Git repository, separate from the NGINX Helm release:

    ```yaml
    # ingressclass.yaml
    apiVersion: networking.k8s.io/v1
    kind: IngressClass
    metadata:
      name: nginx
    spec:
      controller: k8s.io/ingress-nginx
    ```

??? note "If NGINX Was Installed Manually"

    Create the IngressClass as a standalone resource:

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

### Delete NGINX Admission Webhook

You should delete the admission webhook to avoid issues with Ingress modifications after NGINX is removed:

```bash
kubectl delete validatingwebhookconfiguration ingress-nginx-admission
kubectl delete mutatingwebhookconfiguration ingress-nginx-admission --ignore-not-found
```

### Uninstall NGINX

```bash
helm uninstall ingress-nginx -n ingress-nginx
```

If you added the `helm.sh/resource-policy: keep` annotation, you should see:

```text
These resources were kept due to the resource policy:
[IngressClass] nginx

release "ingress-nginx" uninstalled
```

### Verify IngressClass Exists

```bash
kubectl get ingressclass nginx
```

In case, the ingressClass is somehow deleted, you can recreate it using the commands in [Preserve the IngressClass](#preserve-the-ingressclass).

### Clean Up NGINX Namespace

```bash
kubectl delete namespace ingress-nginx
```

!!! success "Migration Complete"

    Congratulations! You have successfully migrated from Ingress NGINX Controller to Traefik with zero downtime. Your existing Ingresses with `ingressClassName: nginx` continue to work, now served by Traefik.

---

## Troubleshooting

There is a dashboard available in Traefik that can help to understand what's going on.
Refer to the [dedicated documentation](../reference/install-configuration/api-dashboard.md#configuration-example) to enable it.

??? note "Ingresses Not Discovered by Traefik"

    ```bash
    # Verify IngressClass exists
    kubectl get ingressclass nginx

    # Check Traefik provider configuration
    kubectl logs -n traefik deployment/traefik | grep -i "nginx\|ingress"

    # Verify Ingress has correct ingressClassName
    kubectl get ingress <name> -o yaml | grep ingressClassName
    ```

??? note "Annotation Not Working as Expected"

    Some NGINX annotations have behavioral differences in Traefik. Check the [limitations documentation](../reference/routing-configuration/kubernetes/ingress-nginx.md#limitations) for details.

??? note "TLS Certificates Not Working"

    Existing TLS configurations continue to work with Traefik:

    - Keep `spec.tls` entries exactly as-is; Traefik terminates TLS using the referenced secrets
    - TLS secrets must stay in the same namespace as the Ingress
    - NGINX `ssl-redirect` / `force-ssl-redirect` annotations are honored

    ```bash
        # Verify TLS secret exists in the same namespace as Ingress
    kubectl get secrets -n <namespace>

        # Check secret format
    kubectl get secret <tls-secret-name> -n <namespace> -o yaml
    ```

??? note "LoadBalancer IP Not Assigned"

    ```bash
      # Check service status
      kubectl describe svc -n traefik traefik

      # Check for events
      kubectl get events -n traefik --sort-by='.lastTimestamp'
    ```

---

## Next Steps

**Learn More About Traefik:**

- [Kubernetes Ingress NGINX Install Configuration](../reference/install-configuration/providers/kubernetes/kubernetes-ingress-nginx.md) - Detailed provider configuration
- [Kubernetes Ingress NGINX Routing Configuration](../reference/routing-configuration/kubernetes/ingress-nginx.md) - Routing rules and annotation support
- [HTTP Middlewares](../reference/routing-configuration/http/middlewares/overview.md) - Extend functionality beyond NGINX annotations
- [TLS Configuration](../reference/routing-configuration/http/tls/overview.md) - Advanced TLS and certificate management

**Enhance Your Setup:**

- Enable [metrics](../reference/install-configuration/observability/metrics.md) and [tracing](../reference/install-configuration/observability/tracing.md)
- Configure [access logs](../reference/install-configuration/observability/logs-and-accesslogs.md) for observability
- Explore [Traefik Middlewares](../reference/routing-configuration/http/middlewares/overview.md) for advanced traffic management
- Migrate from Nginx-based config to Traefik [IngressRoute](../reference/routing-configuration/kubernetes/crd/http/ingressroute.md) or [Kubernetes Gateway API](../reference/routing-configuration/kubernetes/gateway-api.md)
- Consider [Traefik Hub](https://traefik.io/traefik-hub/) for enterprise features like AI & API Gateway, API Management, and advanced security

---

## Feedback and Support

If you encounter issues during migration or have suggestions for improving this guide:

- **Report Issues:** [GitHub Issues](https://github.com/traefik/traefik/issues)
- **Community Support:** [Traefik Community Forum](https://community.traefik.io/)
- **Enterprise Support:** [Traefik Labs Commercial Support](https://traefik.io/pricing/)

We welcome contributions to improve this migration guide. See our [contribution guidelines](../contributing/submitting-pull-requests.md) to get started.
