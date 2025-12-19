# AWS OpenSearch Service Setup Guide

This guide walks you through setting up AWS OpenSearch Service for vector search with weave-cli.

## Prerequisites

- AWS account with billing enabled
- OpenAI API key for generating embeddings
- Basic knowledge of AWS Console or AWS CLI

## Quick Start

### 1. Create OpenSearch Domain

1. Sign in to [AWS Console](https://console.aws.amazon.com/)
2. Navigate to **Amazon OpenSearch Service**
3. Click **Create domain**
4. Configure domain:

#### Domain Name
- **Domain name**: `weave-cli-opensearch` (or your choice)
  - Must be lowercase, 3-28 characters
  - Can include letters, numbers, and hyphens

#### Deployment Type
Choose based on your needs:
- **Development and testing**: 1 node, no standby
- **Production**: 3 nodes with Multi-AZ, standby enabled
- **Serverless**: On-demand (higher cost, fully managed)

**Recommendation for development**: Standard deployment with 1 node

#### Engine Version
- **OpenSearch**: Latest version (2.11+ recommended)
  - Ensure k-NN plugin is supported

#### Instance Type
Choose based on workload:
- **t3.small.search**: Development/testing (~$30/month)
- **t3.medium.search**: Light production (~$60/month)
- **r6g.large.search**: Production with better performance (~$150/month)

**Recommendation**: Start with `t3.small.search`

#### Storage
- **EBS storage type**: General Purpose (SSD) - gp3
- **EBS storage size**: 10 GB minimum, 20-50 GB recommended
- **EBS throughput**: 125 MiB/s (default)
- **EBS IOPS**: 3000 (default for gp3)

### 2. Configure Network

#### Network Options
Two approaches:

**Option A: VPC Access (More Secure)**
- **VPC**: Select your VPC
- **Subnets**: Choose 1-3 subnets (Multi-AZ for production)
- **Security Groups**: Create new or use existing
  - Allow inbound HTTPS (443) from your IP or VPC CIDR

**Option B: Public Access (Simpler for Development)**
- **Enable fine-grained access control**: Yes (required for public access)
- **Access policy**: Configure domain access policy
- No VPC required

**Recommendation for weave-cli**: Public access with fine-grained access control

### 3. Security Configuration

#### Fine-Grained Access Control
**Enable fine-grained access control**: ✅ Yes (required for security)

**Create master user**:
- **Username**: `admin` (or your choice)
- **Password**: Create strong password (min 8 characters, mixed case, numbers, symbols)
- **Confirm password**: Re-enter password

⚠️ **Important**: Save these credentials securely! You'll need them for weave-cli.

Example credentials:
```
Username: admin
Password: MySecureP@ss123!
```

#### Encryption
- **Encryption at rest**: Enabled (recommended)
- **Node-to-node encryption**: Enabled (recommended)
- **Require HTTPS**: Enabled (recommended)

### 4. Access Policy

#### Domain Access Policy
Choose access policy type:

**For Public Access (Development)**:
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "es:*",
      "Resource": "arn:aws:es:us-east-1:123456789012:domain/weave-cli-opensearch/*"
    }
  ]
}
```

**For VPC Access (Production)**:
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "es:*",
      "Resource": "arn:aws:es:us-east-1:123456789012:domain/weave-cli-opensearch/*",
      "Condition": {
        "IpAddress": {
          "aws:SourceIp": [
            "YOUR_IP_ADDRESS/32"
          ]
        }
      }
    }
  ]
}
```

Replace:
- `123456789012` with your AWS account ID
- `us-east-1` with your region
- `YOUR_IP_ADDRESS` with your public IP

### 5. Additional Settings

#### Logs (Optional but Recommended)
Enable CloudWatch Logs for monitoring:
- **Audit logs**: For security auditing
- **Error logs**: For troubleshooting
- **Slow logs**: For performance tuning
  - **Search slow logs**: Enabled
  - **Index slow logs**: Enabled

#### Auto-Tune (Recommended)
- **Enable Auto-Tune**: Yes
- Automatically adjusts cluster settings for optimal performance

#### Advanced Cluster Settings (Keep Defaults)
- **Dedicated master nodes**: No (for t3.small, not needed)
- **Automated snapshots**: Enabled (daily backups)
- **Custom endpoint**: Not needed for weave-cli

### 6. Review and Create

1. Review all settings
2. Estimated cost: Check monthly estimate
3. Click **Create** domain
4. Wait for domain creation (~10-20 minutes)

### 7. Get Endpoint URL

Once domain is active:

1. Navigate to your domain in OpenSearch Service console
2. Copy **Domain endpoint**
3. Format: `https://search-weave-cli-opensearch-abc123xyz.us-east-1.es.amazonaws.com`

Example:
```
Domain endpoint: https://search-weave-cli-opensearch-abc123xyz.us-east-1.es.amazonaws.com
```

## Configure weave-cli

### Method 1: Environment Variables

```bash
export OPENSEARCH_CLOUD_ADDRESS="https://search-weave-cli-opensearch-abc123xyz.us-east-1.es.amazonaws.com"
export OPENSEARCH_CLOUD_USERNAME="admin"
export OPENSEARCH_CLOUD_PASSWORD="MySecureP@ss123!"
export VECTOR_DB_TYPE="opensearch-cloud"
export OPENAI_API_KEY="sk-..."  # For embeddings
```

### Method 2: Configuration File

Create `configs/config.opensearch-cloud.yaml`:

```yaml
vector_db:
  type: opensearch-cloud
  url: "https://search-weave-cli-opensearch-abc123xyz.us-east-1.es.amazonaws.com"
  username: "admin"
  password: "MySecureP@ss123!"
  timeout: 60
  vector_dimensions: 1536
  similarity_metric: cosine
```

### Method 3: Interactive Configuration

```bash
weave config create --env --opensearch-cloud
```

Follow prompts to enter:
- Domain endpoint URL
- Master username
- Master password

## Verify Connection

Test connectivity:

```bash
weave health check --opensearch-cloud
```

Expected output:
```
✓ Connected to OpenSearch Cloud
✓ Cluster status: green
✓ Version: OpenSearch 2.11.0
```

## Usage Examples

### Create Collection with Vector Search

```bash
# Create collection for documents
weave cols create my_documents --opensearch-cloud

# Verify collection exists
weave cols ls --opensearch-cloud
```

### Add Documents

```bash
# Create document from text file
weave docs create my_documents document.txt --opensearch-cloud

# List documents
weave docs ls my_documents --opensearch-cloud
```

## Troubleshooting

### Connection Issues

**Error**: `Connection refused` or `connection timeout`

**Common causes**:
1. Domain not active yet (check AWS Console - status should be "Active")
2. Incorrect endpoint URL (verify format: `https://search-...`)
3. VPC security group blocking access
4. Public access policy not configured

**Solutions**:
- Wait for domain to finish creating (10-20 minutes)
- Verify endpoint URL in AWS Console
- Check security group rules (allow HTTPS/443)
- Review domain access policy

### Authentication Errors

**Error**: `401 Unauthorized` or `403 Forbidden`

**Common causes**:
1. Incorrect username or password
2. Fine-grained access control misconfigured
3. Access policy denying requests
4. IP address not whitelisted

**Solutions**:
- Verify username (default: `admin`)
- Reset master user password in AWS Console:
  - Domain → Security configuration → Edit
  - Master user → Reset password
- Update access policy to allow your IP
- Check CloudWatch logs for detailed error messages

### Performance Issues

**Symptoms**: Slow queries, timeouts, high latency

**Common causes**:
1. Instance type too small (t3.small insufficient for production)
2. Too many documents for instance size
3. Not enough shards
4. Storage I/O bottleneck

**Solutions**:
- Upgrade instance type: t3.small → t3.medium → r6g.large
- Increase shard count when creating indices
- Use faster storage: gp2 → gp3 with higher IOPS
- Enable Auto-Tune for automatic optimization
- Monitor CloudWatch metrics:
  - CPUUtilization
  - JVMMemoryPressure
  - SearchLatency

### Cluster Health Red

**Error**: `cluster status is red`

**Causes**:
- Node failure
- Shard allocation failure
- Insufficient disk space
- Memory pressure

**Solutions**:
1. Check cluster health API:
   ```bash
   curl -u admin:password https://your-endpoint/_cluster/health
   ```
2. Review CloudWatch logs
3. Check disk space utilization (keep below 85%)
4. Increase instance size if JVM memory pressure > 75%
5. Contact AWS Support if issue persists

### Cost Management

**Unexpected high costs**:

**Common causes**:
1. Instance running 24/7 (not needed for dev)
2. Instance type too large
3. Excessive data transfer
4. CloudWatch logs accumulating

**Solutions**:
- Delete domain when not in use (development)
- Use t3.small for development
- Monitor AWS Cost Explorer
- Set up billing alerts
- Disable unnecessary CloudWatch logs

## Advanced Configuration

### Custom k-NN Index Settings

Create collection with custom vector settings:

```json
{
  "settings": {
    "index.knn": true,
    "index.knn.algo_param.ef_search": 100
  },
  "mappings": {
    "properties": {
      "embedding": {
        "type": "knn_vector",
        "dimension": 1536,
        "method": {
          "name": "hnsw",
          "space_type": "cosinesimil",
          "engine": "nmslib",
          "parameters": {
            "ef_construction": 128,
            "m": 16
          }
        }
      }
    }
  }
}
```

### IAM Authentication (Coming Soon)

AWS Signature V4 authentication for enhanced security:

```bash
# Future support
export AWS_ACCESS_KEY_ID="AKIAIOSFODNN7EXAMPLE"
export AWS_SECRET_ACCESS_KEY="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
export AWS_REGION="us-east-1"
```

Note: Currently use username/password. IAM auth support planned.

## Resources

- [AWS OpenSearch Service](https://aws.amazon.com/opensearch-service/)
- [Developer Guide](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/)
- [k-NN Plugin Documentation](https://opensearch.org/docs/latest/search-plugins/knn/)
- [Fine-Grained Access Control](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/fgac.html)
- [AWS Pricing Calculator](https://calculator.aws/#/)

## Support

If you encounter issues:
1. Check [Troubleshooting](#troubleshooting) section above
2. Review AWS OpenSearch CloudWatch logs
3. Check AWS Service Health Dashboard
4. Contact AWS Support (if you have support plan)
5. File an issue on [GitHub](https://github.com/maximilien/weave-cli/issues)
