# Cloud Storage for Snapshots

Kubeshark can upload and download snapshots to cloud object storage, enabling cross-cluster sharing, backup/restore, and long-term retention.

Supported providers: **Amazon S3** (`s3`) and **Azure Blob Storage** (`azblob`).

## Helm Values

```yaml
tap:
  snapshots:
    cloud:
      provider: ""      # "s3" or "azblob" (empty = disabled)
      configMaps: []    # names of pre-existing ConfigMaps with cloud config env vars
      secrets: []       # names of pre-existing Secrets with cloud credentials
```

- `provider` selects which cloud backend to use. Leave empty to disable cloud storage.
- `configMaps` and `secrets` are lists of names of existing ConfigMap/Secret resources. They are mounted as `envFrom` on the hub pod, injecting all their keys as environment variables.

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
