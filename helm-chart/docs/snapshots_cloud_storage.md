# Cloud Storage for Snapshots

Kubeshark can upload and download snapshots to cloud object storage, enabling cross-cluster sharing, backup/restore, and long-term retention.

Supported providers: **Amazon S3** (`s3`), **Azure Blob Storage** (`azblob`), and **Google Cloud Storage** (`gcs`).

## Helm Values

```yaml
tap:
  snapshots:
    cloud:
      provider: ""      # "s3", "azblob", or "gcs" (empty = disabled)
      prefix: ""        # key prefix in the bucket/container (e.g. "snapshots/")
      configMaps: []    # names of pre-existing ConfigMaps with cloud config env vars
      secrets: []       # names of pre-existing Secrets with cloud credentials
      s3:
        bucket: ""
        region: ""
        accessKey: ""
        secretKey: ""
        roleArn: ""
        externalId: ""
      azblob:
        storageAccount: ""
        container: ""
        storageKey: ""
      gcs:
        bucket: ""
        project: ""
        credentialsJson: ""
```

- `provider` selects which cloud backend to use. Leave empty to disable cloud storage.
- `configMaps` and `secrets` are lists of names of existing ConfigMap/Secret resources. They are mounted as `envFrom` on the hub pod, injecting all their keys as environment variables.

### Inline Values (Alternative to External ConfigMaps/Secrets)

Instead of creating ConfigMap and Secret resources manually, you can set cloud storage configuration directly in `values.yaml` or via `--set` flags. The Helm chart will automatically create the necessary ConfigMap and Secret resources.

Both approaches can be used together — inline values are additive to external `configMaps`/`secrets` references.

---

## Amazon S3

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `SNAPSHOT_AWS_BUCKET` | Yes | S3 bucket name |
| `SNAPSHOT_AWS_REGION` | No | AWS region (uses SDK default if empty) |
| `SNAPSHOT_AWS_ACCESS_KEY` | No | Static access key ID (empty = use default credential chain) |
| `SNAPSHOT_AWS_SECRET_KEY` | No | Static secret access key |
| `SNAPSHOT_AWS_ROLE_ARN` | No | IAM role ARN to assume via STS (for cross-account access) |
| `SNAPSHOT_AWS_EXTERNAL_ID` | No | External ID for the STS AssumeRole call |
| `SNAPSHOT_CLOUD_PREFIX` | No | Key prefix in the bucket (e.g. `snapshots/`) |

### Authentication Methods

Credentials are resolved in this order:

1. **Static credentials** -- If `SNAPSHOT_AWS_ACCESS_KEY` is set, static credentials are used directly.
2. **STS AssumeRole** -- If `SNAPSHOT_AWS_ROLE_ARN` is also set, the static (or default) credentials are used to assume the given IAM role. This is useful for cross-account S3 access.
3. **AWS default credential chain** -- When no static credentials are provided, the SDK default chain is used:
   - **IRSA** (EKS service account token) -- recommended for production on EKS
   - EC2 instance profile
   - Standard AWS environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, etc.)
   - Shared credentials file (`~/.aws/credentials`)

The provider validates bucket access on startup via `HeadBucket`. If the bucket is inaccessible, the hub will fail to start.

### Example: Inline Values (simplest approach)

```yaml
tap:
  snapshots:
    cloud:
      provider: "s3"
      s3:
        bucket: my-kubeshark-snapshots
        region: us-east-1
```

Or with static credentials via `--set`:

```bash
helm install kubeshark kubeshark/kubeshark \
  --set tap.snapshots.cloud.provider=s3 \
  --set tap.snapshots.cloud.s3.bucket=my-kubeshark-snapshots \
  --set tap.snapshots.cloud.s3.region=us-east-1 \
  --set tap.snapshots.cloud.s3.accessKey=AKIA... \
  --set tap.snapshots.cloud.s3.secretKey=wJal...
```

### Example: IRSA (recommended for EKS)

Create a ConfigMap with bucket configuration:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: kubeshark-s3-config
data:
  SNAPSHOT_AWS_BUCKET: my-kubeshark-snapshots
  SNAPSHOT_AWS_REGION: us-east-1
```

Set Helm values:

```yaml
tap:
  snapshots:
    cloud:
      provider: "s3"
      configMaps:
        - kubeshark-s3-config
```

The hub pod's service account must be annotated for IRSA with an IAM role that has S3 access to the bucket.

### Example: Static Credentials

Create a Secret with credentials:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: kubeshark-s3-creds
type: Opaque
stringData:
  SNAPSHOT_AWS_ACCESS_KEY: AKIA...
  SNAPSHOT_AWS_SECRET_KEY: wJal...
```

Create a ConfigMap with bucket configuration:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: kubeshark-s3-config
data:
  SNAPSHOT_AWS_BUCKET: my-kubeshark-snapshots
  SNAPSHOT_AWS_REGION: us-east-1
```

Set Helm values:

```yaml
tap:
  snapshots:
    cloud:
      provider: "s3"
      configMaps:
        - kubeshark-s3-config
      secrets:
        - kubeshark-s3-creds
```

### Example: Cross-Account Access via AssumeRole

Add the role ARN to your ConfigMap:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: kubeshark-s3-config
data:
  SNAPSHOT_AWS_BUCKET: other-account-bucket
  SNAPSHOT_AWS_REGION: eu-west-1
  SNAPSHOT_AWS_ROLE_ARN: arn:aws:iam::123456789012:role/KubesharkCrossAccountRole
  SNAPSHOT_AWS_EXTERNAL_ID: my-external-id   # optional, if required by the trust policy
```

The hub will first authenticate using its own credentials (IRSA, static, or default chain), then assume the specified role to access the bucket.

---

## Azure Blob Storage

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `SNAPSHOT_AZBLOB_STORAGE_ACCOUNT` | Yes | Azure storage account name |
| `SNAPSHOT_AZBLOB_CONTAINER` | Yes | Blob container name |
| `SNAPSHOT_AZBLOB_STORAGE_KEY` | No | Storage account access key (empty = use DefaultAzureCredential) |
| `SNAPSHOT_CLOUD_PREFIX` | No | Key prefix in the container (e.g. `snapshots/`) |

### Authentication Methods

Credentials are resolved in this order:

1. **Shared Key** -- If `SNAPSHOT_AZBLOB_STORAGE_KEY` is set, the storage account key is used directly.
2. **DefaultAzureCredential** -- When no storage key is provided, the Azure SDK default credential chain is used:
   - **Workload Identity** (AKS pod identity) -- recommended for production on AKS
   - Managed Identity (system or user-assigned)
   - Azure CLI credentials
   - Environment variables (`AZURE_CLIENT_ID`, `AZURE_TENANT_ID`, `AZURE_CLIENT_SECRET`)

The provider validates container access on startup via `GetProperties`. If the container is inaccessible, the hub will fail to start.

### Example: Inline Values

```yaml
tap:
  snapshots:
    cloud:
      provider: "azblob"
      azblob:
        storageAccount: mykubesharksa
        container: snapshots
        storageKey: "base64-encoded-storage-key..."  # optional, omit for DefaultAzureCredential
```

### Example: Workload Identity (recommended for AKS)

Create a ConfigMap with storage configuration:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: kubeshark-azblob-config
data:
  SNAPSHOT_AZBLOB_STORAGE_ACCOUNT: mykubesharksa
  SNAPSHOT_AZBLOB_CONTAINER: snapshots
```

Set Helm values:

```yaml
tap:
  snapshots:
    cloud:
      provider: "azblob"
      configMaps:
        - kubeshark-azblob-config
```

The hub pod's service account must be configured for AKS Workload Identity with a managed identity that has the **Storage Blob Data Contributor** role on the container.

### Example: Storage Account Key

Create a Secret with the storage key:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: kubeshark-azblob-creds
type: Opaque
stringData:
  SNAPSHOT_AZBLOB_STORAGE_KEY: "base64-encoded-storage-key..."
```

Create a ConfigMap with storage configuration:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: kubeshark-azblob-config
data:
  SNAPSHOT_AZBLOB_STORAGE_ACCOUNT: mykubesharksa
  SNAPSHOT_AZBLOB_CONTAINER: snapshots
```

Set Helm values:

```yaml
tap:
  snapshots:
    cloud:
      provider: "azblob"
      configMaps:
        - kubeshark-azblob-config
      secrets:
        - kubeshark-azblob-creds
```

---

## Google Cloud Storage

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `SNAPSHOT_GCS_BUCKET` | Yes | GCS bucket name |
| `SNAPSHOT_GCS_PROJECT` | No | GCP project ID |
| `SNAPSHOT_GCS_CREDENTIALS_JSON` | No | Service account JSON key (empty = use Application Default Credentials) |
| `SNAPSHOT_CLOUD_PREFIX` | No | Key prefix in the bucket (e.g. `snapshots/`) |

### Authentication Methods

Credentials are resolved in this order:

1. **Service Account JSON Key** -- If `SNAPSHOT_GCS_CREDENTIALS_JSON` is set, the provided JSON key is used directly.
2. **Application Default Credentials** -- When no JSON key is provided, the GCP SDK default credential chain is used:
   - **Workload Identity** (GKE pod identity) -- recommended for production on GKE
   - GCE instance metadata (Compute Engine default service account)
   - Standard GCP environment variables (`GOOGLE_APPLICATION_CREDENTIALS`)
   - `gcloud` CLI credentials

The provider validates bucket access on startup via `Bucket.Attrs()`. If the bucket is inaccessible, the hub will fail to start.

### Required IAM Permissions

The service account needs different IAM roles depending on the access level:

**Read-only** (download, list, and sync snapshots from cloud):

| Role | Permissions provided | Purpose |
|------|---------------------|---------|
| `roles/storage.legacyBucketReader` | `storage.buckets.get`, `storage.objects.list` | Hub startup (bucket validation) + listing snapshots |
| `roles/storage.objectViewer` | `storage.objects.get`, `storage.objects.list` | Downloading snapshots, checking existence, reading metadata |

```bash
gcloud storage buckets add-iam-policy-binding gs://BUCKET_NAME \
  --member="serviceAccount:SA_EMAIL" \
  --role="roles/storage.legacyBucketReader"
gcloud storage buckets add-iam-policy-binding gs://BUCKET_NAME \
  --member="serviceAccount:SA_EMAIL" \
  --role="roles/storage.objectViewer"
```

**Read-write** (upload and delete snapshots in addition to read):

Add `roles/storage.objectAdmin` instead of `roles/storage.objectViewer` to also grant `storage.objects.create` and `storage.objects.delete`:

| Role | Permissions provided | Purpose |
|------|---------------------|---------|
| `roles/storage.legacyBucketReader` | `storage.buckets.get`, `storage.objects.list` | Hub startup (bucket validation) + listing snapshots |
| `roles/storage.objectAdmin` | `storage.objects.*` | Full object CRUD (upload, download, delete, list, metadata) |

```bash
gcloud storage buckets add-iam-policy-binding gs://BUCKET_NAME \
  --member="serviceAccount:SA_EMAIL" \
  --role="roles/storage.legacyBucketReader"
gcloud storage buckets add-iam-policy-binding gs://BUCKET_NAME \
  --member="serviceAccount:SA_EMAIL" \
  --role="roles/storage.objectAdmin"
```

### Example: Inline Values (simplest approach)

```yaml
tap:
  snapshots:
    cloud:
      provider: "gcs"
      gcs:
        bucket: my-kubeshark-snapshots
        project: my-gcp-project
```

Or with a service account key via `--set`:

```bash
helm install kubeshark kubeshark/kubeshark \
  --set tap.snapshots.cloud.provider=gcs \
  --set tap.snapshots.cloud.gcs.bucket=my-kubeshark-snapshots \
  --set tap.snapshots.cloud.gcs.project=my-gcp-project \
  --set-file tap.snapshots.cloud.gcs.credentialsJson=service-account.json
```

### Example: Workload Identity (recommended for GKE)

Create a ConfigMap with bucket configuration:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: kubeshark-gcs-config
data:
  SNAPSHOT_GCS_BUCKET: my-kubeshark-snapshots
  SNAPSHOT_GCS_PROJECT: my-gcp-project
```

Set Helm values:

```yaml
tap:
  snapshots:
    cloud:
      provider: "gcs"
      configMaps:
        - kubeshark-gcs-config
```

Configure GKE Workload Identity to allow the Kubernetes service account to impersonate the GCP service account:

```bash
# Ensure the GKE cluster has Workload Identity enabled
# (--workload-pool=PROJECT_ID.svc.id.goog at cluster creation)

# Create a GCP service account (if not already created)
gcloud iam service-accounts create kubeshark-gcs \
  --display-name="Kubeshark GCS Snapshots"

# Grant bucket access (read-write — see Required IAM Permissions above)
gcloud storage buckets add-iam-policy-binding gs://BUCKET_NAME \
  --member="serviceAccount:kubeshark-gcs@PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/storage.legacyBucketReader"
gcloud storage buckets add-iam-policy-binding gs://BUCKET_NAME \
  --member="serviceAccount:kubeshark-gcs@PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/storage.objectAdmin"

# Allow the K8s service account to impersonate the GCP service account
# Note: the K8s SA name is "<release-name>-service-account" (default: "kubeshark-service-account")
gcloud iam service-accounts add-iam-policy-binding \
  kubeshark-gcs@PROJECT_ID.iam.gserviceaccount.com \
  --role="roles/iam.workloadIdentityUser" \
  --member="serviceAccount:PROJECT_ID.svc.id.goog[NAMESPACE/kubeshark-service-account]"
```

Set Helm values — the `tap.annotations` field adds the Workload Identity annotation to the service account:

```yaml
tap:
  annotations:
    iam.gke.io/gcp-service-account: kubeshark-gcs@PROJECT_ID.iam.gserviceaccount.com
  snapshots:
    cloud:
      provider: "gcs"
      configMaps:
        - kubeshark-gcs-config
```

Or via `--set`:

```bash
helm install kubeshark kubeshark/kubeshark \
  --set tap.snapshots.cloud.provider=gcs \
  --set tap.snapshots.cloud.gcs.bucket=BUCKET_NAME \
  --set tap.snapshots.cloud.gcs.project=PROJECT_ID \
  --set tap.annotations."iam\.gke\.io/gcp-service-account"=kubeshark-gcs@PROJECT_ID.iam.gserviceaccount.com
```

No `credentialsJson` secret is needed — GKE injects credentials automatically via the Workload Identity metadata server.

### Example: Service Account Key

Create a Secret with the service account JSON key:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: kubeshark-gcs-creds
type: Opaque
stringData:
  SNAPSHOT_GCS_CREDENTIALS_JSON: |
    {
      "type": "service_account",
      "project_id": "my-gcp-project",
      "private_key_id": "...",
      "private_key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n",
      "client_email": "kubeshark@my-gcp-project.iam.gserviceaccount.com",
      ...
    }
```

Create a ConfigMap with bucket configuration:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: kubeshark-gcs-config
data:
  SNAPSHOT_GCS_BUCKET: my-kubeshark-snapshots
  SNAPSHOT_GCS_PROJECT: my-gcp-project
```

Set Helm values:

```yaml
tap:
  snapshots:
    cloud:
      provider: "gcs"
      configMaps:
        - kubeshark-gcs-config
      secrets:
        - kubeshark-gcs-creds
```
