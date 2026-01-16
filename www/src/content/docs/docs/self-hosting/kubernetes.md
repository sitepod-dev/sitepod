---
title: Kubernetes
description: Deploy SitePod on Kubernetes
---

Deploy SitePod to a Kubernetes cluster with automatic SSL via an Ingress controller.

## Prerequisites

- Kubernetes cluster
- kubectl configured
- Ingress controller (nginx-ingress or similar)
- cert-manager (for automatic TLS)

## Manifests

### ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: sitepod-caddyfile
data:
  Caddyfile: |
    {
        admin off
        auto_https off
        order sitepod first
    }
    :8080 {
        sitepod {
            storage_path /data
            data_dir /data
            domain sitepod.example.com
        }
    }
```

### Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: sitepod-credentials
type: Opaque
stringData:
  admin-email: admin@example.com
  admin-password: YourSecurePassword123
```

### PersistentVolumeClaim

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: sitepod-data
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 20Gi
```

### Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sitepod
spec:
  replicas: 1
  selector:
    matchLabels:
      app: sitepod
  template:
    metadata:
      labels:
        app: sitepod
    spec:
      containers:
      - name: sitepod
        image: ghcr.io/sitepod-dev/sitepod:latest
        ports:
        - containerPort: 8080
        env:
        - name: SITEPOD_DOMAIN
          value: sitepod.example.com
        - name: SITEPOD_ADMIN_EMAIL
          valueFrom:
            secretKeyRef:
              name: sitepod-credentials
              key: admin-email
        - name: SITEPOD_ADMIN_PASSWORD
          valueFrom:
            secretKeyRef:
              name: sitepod-credentials
              key: admin-password
        volumeMounts:
        - name: data
          mountPath: /data
        - name: caddyfile
          mountPath: /etc/caddy/Caddyfile
          subPath: Caddyfile
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
          limits:
            memory: "1Gi"
            cpu: "1000m"
        livenessProbe:
          httpGet:
            path: /api/v1/health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /api/v1/health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: sitepod-data
      - name: caddyfile
        configMap:
          name: sitepod-caddyfile
```

### Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: sitepod
spec:
  selector:
    app: sitepod
  ports:
  - port: 80
    targetPort: 8080
```

### Ingress

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: sitepod
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - "sitepod.example.com"
    - "*.sitepod.example.com"
    secretName: sitepod-tls
  rules:
  - host: "sitepod.example.com"
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: sitepod
            port:
              number: 80
  - host: "*.sitepod.example.com"
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: sitepod
            port:
              number: 80
```

## Deployment

Apply all manifests:

```bash
kubectl apply -f sitepod-configmap.yaml
kubectl apply -f sitepod-secret.yaml
kubectl apply -f sitepod-pvc.yaml
kubectl apply -f sitepod-deployment.yaml
kubectl apply -f sitepod-service.yaml
kubectl apply -f sitepod-ingress.yaml
```

## Wildcard certificates

For wildcard certificates, you need DNS-01 challenge. Example with Cloudflare:

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@example.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - dns01:
        cloudflare:
          apiTokenSecretRef:
            name: cloudflare-api-token
            key: api-token
```

## Scaling considerations

:::note
SitePod uses SQLite, which requires single-writer access. Running multiple replicas is not supported without external storage.
:::

For high availability, consider:
- Using S3/R2 for blob storage
- Running a single SitePod instance with proper resource limits
- Using node affinity to ensure consistent scheduling

## Next steps

- [Storage Backends](/docs/self-hosting/storage/) - S3 for Kubernetes
- [SSL/TLS Options](/docs/self-hosting/ssl/) - Certificate configuration
