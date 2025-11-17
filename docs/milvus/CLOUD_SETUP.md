# Milvus Cloud Setup Guide (Zilliz)

This guide covers setting up Milvus with Zilliz Cloud, the fully managed Milvus service.

## Overview

Zilliz Cloud provides:

- **Managed Infrastructure**: No server management required
- **High Availability**: Automatic failover and redundancy
- **Scalability**: Elastic scaling based on demand
- **Security**: Enterprise-grade security and compliance
- **Global Deployment**: Multiple regions worldwide
- **Monitoring**: Built-in observability and alerting

## Prerequisites

1. **Zilliz Cloud Account**: Sign up at [cloud.zilliz.com](https://cloud.zilliz.com)
2. **API Credentials**: Cluster endpoint, username, and password
3. **OpenAI API Key**: For automatic embedding generation (optional)

## Quick Start

### 1. Create a Zilliz Cluster

1. **Sign up** at [cloud.zilliz.com](https://cloud.zilliz.com)
2. **Create a new cluster**:
   - Choose region (us-west-2, eu-west-1, ap-southeast-1, etc.)
   - Select plan (Free tier available for testing)
   - Choose cluster size (based on your data volume)
3. **Get credentials**:
   - Cluster endpoint: `your-cluster.aws-us-west-2.vectordb.zillizcloud.com:19530`
   - Username: Provided during cluster creation
   - Password: Set during cluster creation

### 2. Configure Environment

```bash
# Set Zilliz credentials
export MILVUS_CLOUD_ADDRESS="your-cluster.aws-us-west-2.vectordb.zillizcloud.com:19530"
export MILVUS_CLOUD_USERNAME="your-username"
export MILVUS_CLOUD_PASSWORD="your-password"

# Set OpenAI key for embeddings
export OPENAI_API_KEY="sk-..."
```

### 3. Test Connection

```bash
# Test connection
./tools/vdb/health.sh milvus-cloud

# List collections
weave --milvus-cloud collections list

# Create test collection
weave --milvus-cloud collections create test_collection
```

## Configuration

### Environment Variables

Create a `.env` file:

```bash
# Zilliz Cloud credentials
MILVUS_CLOUD_ADDRESS="your-cluster.aws-us-west-2.vectordb.zillizcloud.com:19530"
MILVUS_CLOUD_USERNAME="db_admin"
MILVUS_CLOUD_PASSWORD="your-secure-password"

# Database (optional, defaults to "default")
MILVUS_CLOUD_DATABASE="default"

# Timeout for cloud operations (recommended: 30-60 seconds)
MILVUS_CLOUD_TIMEOUT="30"

# OpenAI for embeddings
OPENAI_API_KEY="sk-..."
```

Load environment:

```bash
source .env
```

### Configuration File

Create `config.milvus-cloud.yaml`:

```yaml
vectordb:
  type: milvus-cloud

  # Cluster connection
  address: ${MILVUS_CLOUD_ADDRESS}
  username: ${MILVUS_CLOUD_USERNAME}
  password: ${MILVUS_CLOUD_PASSWORD}

  # Database settings
  database: default
  timeout: 30  # Longer timeout for network latency

  # Vector configuration
  vector_dimensions: 1536  # text-embedding-3-small
  similarity_metric: COSINE  # Recommended for normalized embeddings

# Optional: LLM configuration
llm:
  provider: openai
  api_key: ${OPENAI_API_KEY}
  embedding_model: text-embedding-3-small
```

Use with weave:

```bash
weave --config config.milvus-cloud.yaml --milvus-cloud collections list
```

## Security Best Practices

### 1. Credential Management

**DO NOT** hardcode credentials in code or config files:

```bash
# ❌ Bad: Hardcoded credentials
address: "cluster.zillizcloud.com:19530"
username: "admin"
password: "password123"

# ✅ Good: Environment variables
address: ${MILVUS_CLOUD_ADDRESS}
username: ${MILVUS_CLOUD_USERNAME}
password: ${MILVUS_CLOUD_PASSWORD}
```

### 2. Use Secret Management

**For Production:**

```bash
# AWS Secrets Manager
aws secretsmanager get-secret-value \
  --secret-id prod/milvus/credentials \
  --query SecretString \
  --output text | jq -r .password

# HashiCorp Vault
vault kv get -field=password secret/milvus/prod

# Kubernetes Secrets
kubectl get secret milvus-credentials -o jsonpath='{.data.password}' | base64 -d
```

**In Scripts:**

```bash
#!/bin/bash

# Load from secret manager
MILVUS_CLOUD_PASSWORD=$(aws secretsmanager get-secret-value \
  --secret-id prod/milvus/password \
  --query SecretString \
  --output text)

export MILVUS_CLOUD_PASSWORD

# Run weave
weave --milvus-cloud collections list
```

### 3. IP Allowlisting

Configure IP allowlist in Zilliz Cloud console:

1. Go to **Cluster Settings** → **Network**
2. Add your IP addresses or CIDR ranges
3. Use VPN or bastion host for production access

### 4. TLS/SSL

Zilliz Cloud uses TLS by default:

- Connections are encrypted in transit
- Certificate validation is automatic
- No additional configuration needed

## Cluster Management

### Sizing Recommendations

**Free Tier (Testing):**
- 1 CU (Capacity Unit)
- ~100K vectors
- Good for development and testing

**Starter (Small Production):**
- 2-4 CUs
- ~1M vectors
- Suitable for small applications

**Professional (Medium Production):**
- 8-16 CUs
- ~10M vectors
- Suitable for most production workloads

**Enterprise (Large Production):**
- 32+ CUs
- 100M+ vectors
- High-traffic applications

### Scaling

**Manual Scaling:**
1. Go to Zilliz Cloud console
2. Select cluster → **Settings**
3. Adjust capacity units
4. Apply changes (takes 5-10 minutes)

**Auto-scaling:**
- Available on Professional and Enterprise plans
- Configure min/max CUs
- Set CPU/memory thresholds

### Monitoring

**Zilliz Cloud Console:**
- Query performance metrics
- Resource usage (CPU, memory, disk)
- Request rate and latency
- Error rates

**Custom Monitoring:**

```bash
# Check cluster health
weave --milvus-cloud health

# Get collection statistics
weave --milvus-cloud collections list

# Monitor query performance
weave --milvus-cloud search vector my_collection \
  --query "test" \
  --limit 10 \
  --explain  # Shows query plan and metrics
```

## Usage Examples

### Collection Management

```bash
# Create collection
weave --milvus-cloud collections create products \
  --vectorizer text-embedding-3-small

# List collections
weave --milvus-cloud collections list

# Show collection details
weave --milvus-cloud collections show products

# Delete collection
weave --milvus-cloud collections delete products
```

### Document Operations

```bash
# Add single document
weave --milvus-cloud documents create products \
  --text "Premium wireless headphones with noise cancellation" \
  --metadata '{
    "category": "electronics",
    "price": 299.99,
    "brand": "AudioTech",
    "in_stock": true
  }'

# Add multiple documents from JSON
cat products.json | weave --milvus-cloud documents import products

# List documents
weave --milvus-cloud documents list products --limit 100

# Delete document
weave --milvus-cloud documents delete products doc-12345
```

### Search Operations

**Vector Search:**
```bash
weave --milvus-cloud search vector products \
  --query "wireless noise cancelling headphones" \
  --limit 10 \
  --filter 'price < 500 and in_stock == true'
```

**BM25 Text Search:**
```bash
weave --milvus-cloud search bm25 products \
  --query "premium headphones" \
  --limit 10
```

**Hybrid Search:**
```bash
weave --milvus-cloud search hybrid products \
  --query "affordable bluetooth earbuds" \
  --limit 10
```

## Performance Optimization

### 1. Index Selection

**HNSW** (Recommended for cloud):
```yaml
# Best for high recall and good performance
# Suitable for most use cases
index_type: HNSW
index_params:
  M: 16          # Connections per layer
  efConstruction: 200  # Build time quality
```

**IVF_FLAT** (Default):
```yaml
# Good balance of speed and accuracy
# Suitable for medium datasets
index_type: IVF_FLAT
index_params:
  nlist: 1024    # Adjust based on dataset size
```

**IVF_PQ** (Large datasets):
```yaml
# Memory efficient for large datasets
# Slight accuracy tradeoff
index_type: IVF_PQ
index_params:
  nlist: 2048
  m: 16          # Sub-quantizers
  nbits: 8       # Bits per sub-quantizer
```

### 2. Query Optimization

**Use Filters:**
```bash
# More efficient than post-filtering results
weave --milvus-cloud search vector products \
  --query "laptop" \
  --filter 'price >= 500 and price <= 1500' \
  --limit 20
```

**Batch Queries:**
```bash
# Process multiple queries in one request
weave --milvus-cloud search batch products \
  --queries queries.json \
  --limit 10
```

**Adjust Consistency Level:**
```yaml
# In code (consistency vs latency tradeoff)
consistency_level: "Eventually"  # Faster, eventual consistency
# consistency_level: "Strong"    # Slower, immediate consistency
```

### 3. Data Partitioning

```bash
# Partition by date for time-series data
weave --milvus-cloud collections create logs \
  --partition-key date

# Query specific partition
weave --milvus-cloud search vector logs \
  --partition 2025-01-15 \
  --query "error message" \
  --limit 50
```

## Cost Optimization

### 1. Right-sizing

**Monitor usage:**
```bash
# Check actual usage vs allocated capacity
# Adjust cluster size accordingly
```

**Tips:**
- Start small and scale up as needed
- Use free tier for development
- Monitor query patterns to optimize index

### 2. Data Lifecycle

**Archive old data:**
```bash
# Delete old partitions
weave --milvus-cloud partitions delete logs_2024_12

# Export to cold storage
weave --milvus-cloud collections export archive \
  --output s3://my-bucket/archive/
```

### 3. Query Efficiency

- Use appropriate `limit` values (don't request more than needed)
- Leverage filters to reduce search space
- Use batch operations for multiple queries

## Disaster Recovery

### Backup

**Zilliz Cloud Backups:**
- Automatic daily backups (Professional/Enterprise plans)
- Point-in-time recovery
- Cross-region replication available

**Manual Export:**
```bash
# Export collection
weave --milvus-cloud collections export products \
  --output ./backups/products-$(date +%Y%m%d).json

# Store in S3
aws s3 cp ./backups/products-*.json \
  s3://my-backups/milvus/
```

### Restore

**From Zilliz Backup:**
1. Go to Zilliz Cloud console
2. Select cluster → **Backups**
3. Choose backup point
4. Restore to cluster

**From Export:**
```bash
# Import collection
weave --milvus-cloud collections import products \
  --input ./backups/products-20250115.json
```

## Migration

### From Local to Cloud

```bash
# Export from local
weave --milvus-local collections export products \
  --output products.json

# Import to cloud
weave --milvus-cloud collections import products \
  --input products.json
```

### From Other Vector DBs

```bash
# Generic migration workflow
# 1. Export from source
weave --weaviate collections export products \
  --output products.json

# 2. Transform if needed
python transform_schema.py products.json milvus_products.json

# 3. Import to Milvus cloud
weave --milvus-cloud collections import products \
  --input milvus_products.json
```

## Troubleshooting

### Connection Issues

**Problem:** Cannot connect to cluster

**Solutions:**
```bash
# 1. Verify credentials
echo $MILVUS_CLOUD_ADDRESS
echo $MILVUS_CLOUD_USERNAME
# (don't echo password in production!)

# 2. Check IP allowlist in Zilliz console

# 3. Verify cluster is running
# Check Zilliz Cloud console status

# 4. Test network connectivity
nc -zv $(echo $MILVUS_CLOUD_ADDRESS | cut -d: -f1) 19530
```

### Authentication Errors

**Problem:** Authentication failed

**Solutions:**
```bash
# 1. Verify username and password
# Reset password in Zilliz console if needed

# 2. Check for special characters
# Escape special chars in password: !, $, etc.

# 3. Regenerate credentials
# Create new database user in Zilliz console
```

### Performance Issues

**Problem:** Queries are slow

**Solutions:**

1. **Check cluster resources:**
   - Monitor CPU/memory in Zilliz console
   - Scale up if resources are maxed out

2. **Optimize indexes:**
   - Use HNSW for better performance
   - Adjust nlist for IVF indexes

3. **Reduce result size:**
   ```bash
   # Use smaller limit
   weave --milvus-cloud search vector products \
     --query "laptop" \
     --limit 10  # Instead of 100
   ```

4. **Use filters:**
   ```bash
   # Pre-filter to reduce search space
   weave --milvus-cloud search vector products \
     --query "laptop" \
     --filter 'category == "electronics"' \
     --limit 10
   ```

### Quota Exceeded

**Problem:** Rate limit or quota errors

**Solutions:**

1. **Check plan limits:**
   - Verify QPS limits for your plan
   - Upgrade plan if needed

2. **Implement backoff:**
   ```bash
   # Retry with exponential backoff
   for i in {1..5}; do
     weave --milvus-cloud collections list && break
     sleep $((2**i))
   done
   ```

3. **Batch operations:**
   - Use bulk insert instead of individual inserts
   - Batch queries together

## Support

### Zilliz Cloud Support

- **Documentation**: [docs.zilliz.com](https://docs.zilliz.com)
- **Support Portal**: Available in Zilliz Cloud console
- **Community**: [Discord](https://discord.gg/milvus)
- **Enterprise Support**: Available on Professional/Enterprise plans

### Weave CLI Support

- **GitHub Issues**: [github.com/maximilien/weave-cli/issues](https://github.com/maximilien/weave-cli/issues)
- **Documentation**: [Local Milvus Setup](LOCAL_SETUP.md)

## Next Steps

- [Milvus Documentation](README.md) - General Milvus integration guide
- [Local Setup](LOCAL_SETUP.md) - Set up local development environment
- [Zilliz Best Practices](https://docs.zilliz.com/docs/best-practices) - Official optimization guide
