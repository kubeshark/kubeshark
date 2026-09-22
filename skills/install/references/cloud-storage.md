# Cloud Storage for Snapshots

This is a pointer to the authoritative cloud storage documentation maintained in
the Helm chart:

**Source of truth**: `helm-chart/docs/snapshots_cloud_storage.md`

Always read that file for the latest configuration details, including:

- Amazon S3 (static credentials, IRSA, cross-account AssumeRole)
- Azure Blob Storage (storage key, Workload Identity / DefaultAzureCredential)
- Google Cloud Storage (service account JSON, GKE Workload Identity)
- IAM permissions and trust policy examples
- ConfigMap and Secret setup patterns
- Inline values vs. external ConfigMap/Secret approaches

## Quick Reference

### Helm Values Structure

```yaml
tap:
  snapshots:
    cloud:
      provider: ""      # "s3", "azblob", or "gcs" (empty = disabled)
      prefix: ""        # Key prefix in the bucket/container
      configMaps: []    # Pre-existing ConfigMaps with cloud config env vars
      secrets: []       # Pre-existing Secrets with cloud credentials
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

### Recommended Auth Per Provider

| Provider | Production Recommendation |
|----------|-------------------------|
| S3 (EKS) | IRSA (IAM Roles for Service Accounts) — no static credentials |
| S3 (non-EKS) | Static credentials via Secret, or default AWS credential chain |
| Azure Blob (AKS) | Workload Identity / Managed Identity |
| Azure Blob (non-AKS) | Storage account key via Secret |
| GCS (GKE) | GKE Workload Identity — no JSON key file |
| GCS (non-GKE) | Service account JSON key via Secret |

### Inline Values (Simplest Approach)

Set credentials directly in values.yaml. The Helm chart creates the necessary
ConfigMap/Secret resources automatically.

**S3:**
```yaml
tap:
  snapshots:
    cloud:
      provider: "s3"
      s3:
        bucket: my-kubeshark-snapshots
        region: us-east-1
```

**GCS:**
```yaml
tap:
  snapshots:
    cloud:
      provider: "gcs"
      gcs:
        bucket: my-kubeshark-snapshots
        project: my-gcp-project
```

**Azure Blob:**
```yaml
tap:
  snapshots:
    cloud:
      provider: "azblob"
      azblob:
        storageAccount: mykubesharksa
        container: snapshots
```

For production setups with proper IAM integration, see the full documentation
in `helm-chart/docs/snapshots_cloud_storage.md`.
