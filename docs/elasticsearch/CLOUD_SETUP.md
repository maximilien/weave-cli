# Elasticsearch Cloud Setup (Elastic Cloud)

Complete guide for using Elastic Cloud with Weave CLI.

## What is Elastic Cloud?

Elastic Cloud is the official managed Elasticsearch service by Elastic:
- **Fully managed** - No infrastructure to maintain
- **Auto-scaling** - Scales with your needs
- **Built-in Kibana** - Visualization and monitoring
- **Multi-region** - Deploy globally
- **Free trial** - 14-day trial available

## Prerequisites

- Elastic Cloud account (free trial available)
- Credit card (for trial activation)
- OpenAI API key for embeddings

## Step 1: Create Account

### Sign Up

1. Go to [cloud.elastic.co](https://cloud.elastic.co/registration)
2. Click "Start free trial"
3. Fill in your details
4. Verify email
5. Add payment method (won't be charged during trial)

## Step 2: Create Deployment

### Quick Deployment

1. **Log in** to Elastic Cloud console
2. Click **"Create deployment"**
3. Choose settings:
   - **Name:** `weave-elasticsearch` (or your choice)
   - **Cloud provider:** AWS / GCP / Azure
   - **Region:** Choose closest to you
   - **Version:** 8.11+ (latest)
   - **Hardware profile:** General purpose
   - **Size:** Start with 1GB RAM (can scale later)

4. Click **"Create deployment"**
5. **IMPORTANT:** Copy credentials shown:
   - Username: `elastic`
   - Password: `<random-password>` (save this!)
   - Cloud ID: `deployment:base64string`

### Wait for Deployment

Deployment takes 5-10 minutes. Watch the progress indicator.

## Step 3: Get Credentials

### Method 1: Cloud ID (Recommended)

**Cloud ID** is the easiest way to connect:

1. Go to deployment **"Manage"** page
2. Copy **"Cloud ID"** from deployment overview
3. Looks like: `deployment:dXMtZWFzdC0xLmF3cy5mb3VuZC5pbyQxMjM0NTY=`

### Method 2: Create API Key

**API Keys** are more secure than passwords:

1. In deployment, go to **"Stack Management"**
2. Navigate to **"Security"** > **"API keys"**
3. Click **"Create API key"**
4. Settings:
   - **Name:** `weave-cli`
   - **Expiration:** None (or set expiry)
   - **Restrict privileges:** No (or customize)
5. Copy the **Base64 encoded** API key
6. **Save it** - You can't see it again!

## Step 4: Configure Weave CLI

### Option A: Environment Variables + Config Creator

```bash
# Set credentials
export ELASTICSEARCH_CLOUD_ID="deployment:dXMtZWFzdC0xLmF3cy5mb3VuZC5pbyQxMjM0NTY="
export ELASTICSEARCH_CLOUD_API_KEY="your-base64-api-key"
export OPENAI_API_KEY="sk-..."

# Create config interactively
weave config create --env
# Select: elasticsearch-cloud
```

### Option B: Manual Configuration

Create `configs/config.elasticsearch-cloud.yaml`:

```yaml
databases:
  default: elasticsearch
  vector_databases:
    - name: elasticsearch
      type: elasticsearch-cloud

      # Cloud ID from deployment
      cloud_id: ${ELASTICSEARCH_CLOUD_ID}

      # API key authentication (recommended)
      api_key: ${ELASTICSEARCH_CLOUD_API_KEY}

      # Or basic auth (alternative)
      # username: ${ELASTICSEARCH_CLOUD_USERNAME:-elastic}
      # password: ${ELASTICSEARCH_CLOUD_PASSWORD}

      timeout: 60  # Cloud needs longer timeout
      vector_dimensions: 1536
      similarity_metric: cosine
```

### Option C: Using Password Instead of API Key

```bash
# If using password instead of API key
export ELASTICSEARCH_CLOUD_ID="deployment:..."
export ELASTICSEARCH_CLOUD_USERNAME="elastic"
export ELASTICSEARCH_CLOUD_PASSWORD="your-deployment-password"
export OPENAI_API_KEY="sk-..."
```

## Step 5: Verify Connection

```bash
# Test connection
weave health check

# Expected output:
# ✓ Elasticsearch connection successful
# ✓ Cluster status: green
# ✓ OpenAI API key valid
```

## Step 6: First Operations

```bash
# Create collection
weave cols create MyDocs --text

# Add documents
echo "Hello from Elastic Cloud!" > test.txt
weave docs create MyDocs test.txt

# Search
weave cols q MyDocs "cloud"

# View in Kibana
# Go to: Deployment > Kibana > Dev Tools
# Run: GET /MyDocs/_search
```

## Managing Your Deployment

### Access Kibana

1. Go to deployment page
2. Click **"Kibana"**
3. Log in with `elastic` user
4. Explore your data

### Monitor Usage

**Check cluster health:**
```bash
# Via Weave
weave health check

# Via Kibana Dev Tools
GET /_cluster/health
```

**View metrics:**
- Deployment overview shows CPU, memory, storage
- Set up alerts for resource usage

### Scale Your Deployment

**When to scale:**
- Search getting slow
- High memory usage (>80%)
- Storage nearly full

**How to scale:**
1. Go to deployment **"Edit"**
2. Adjust **"Size per zone"**
3. Click **"Save"**
4. Deployment auto-migrates (no downtime)

### Backup and Restore

**Snapshots are automatic:**
- Hourly snapshots for 24 hours
- Daily snapshots for 7 days
- Weekly snapshots for 4 weeks

**Restore from snapshot:**
1. **Stack Management** > **Snapshot and Restore**
2. Select snapshot
3. Restore indices

## Cost Management

### Free Trial

- **Duration:** 14 days
- **Limits:** 8GB RAM max
- **After trial:** Provide payment or deployment suspends

### Optimize Costs

**Choose right size:**
```
Development:  1-2GB RAM (~$50/month)
Production:   4-8GB RAM (~$200-400/month)
Enterprise:   16GB+ RAM (custom pricing)
```

**Reduce costs:**
- Delete unused indices
- Use lifecycle policies to delete old data
- Scale down during off-hours (if applicable)
- Use reserved instances for discounts

### Monitor Spending

1. **Billing** section shows current usage
2. Set up billing alerts
3. Review monthly invoices

## Troubleshooting

### Connection Timeout

**Cause:** Network/firewall issues

**Fix:**
```bash
# Test connectivity
curl https://your-deployment.es.region.cloud.es.io:9243

# Check firewall allows outbound HTTPS (443)
# Check Cloud ID is correct
echo $ELASTICSEARCH_CLOUD_ID
```

### Authentication Failed

**Cause:** Wrong credentials

**Fix:**
```bash
# Verify Cloud ID
echo $ELASTICSEARCH_CLOUD_ID

# Verify API key is base64 encoded
echo $ELASTICSEARCH_CLOUD_API_KEY

# Or reset elastic user password:
# Deployment > Security > Reset password
```

### Cluster Status Red

**Cause:** Shard allocation issues

**Check:**
```
GET /_cluster/health
GET /_cat/shards?v
```

**Fix:** Usually auto-resolves. Contact Elastic support if persists.

### Out of Storage

**Symptoms:**
- Can't index new documents
- Cluster status yellow/red

**Fix:**
1. Delete old indices
2. Or scale up storage
3. Or set up index lifecycle management

## Security Best Practices

### 1. Use API Keys

**Why:** More secure than passwords, can be scoped

```bash
# Create limited API key for specific indices
POST /_security/api_key
{
  "name": "weave-readonly",
  "role_descriptors": {
    "weave_reader": {
      "indices": [
        {
          "names": ["MyDocs*"],
          "privileges": ["read"]
        }
      ]
    }
  }
}
```

### 2. Network Security

- **IP Filtering:** Enable in deployment settings
- **Private Link:** For enterprise (AWS/GCP/Azure)
- **VPN:** Connect via VPN if needed

### 3. Rotate Credentials

- Rotate API keys every 90 days
- Don't commit keys to git
- Use secret management (AWS Secrets Manager, etc.)

### 4. Monitor Access

- Enable audit logging
- Review access logs in Kibana
- Set up alerts for suspicious activity

## Advanced Features

### Custom Index Templates

```bash
# Create template for vector collections
PUT /_index_template/weave_vectors
{
  "index_patterns": ["weave_*"],
  "template": {
    "settings": {
      "number_of_shards": 1,
      "number_of_replicas": 1
    },
    "mappings": {
      "properties": {
        "vector_field": {
          "type": "dense_vector",
          "dims": 1536,
          "index": true,
          "similarity": "cosine"
        }
      }
    }
  }
}
```

### Index Lifecycle Management

**Auto-delete old data:**
```bash
# Delete indices after 30 days
PUT /_ilm/policy/weave_delete_30d
{
  "policy": {
    "phases": {
      "delete": {
        "min_age": "30d",
        "actions": {
          "delete": {}
        }
      }
    }
  }
}
```

## Next Steps

- [README.md](README.md) - Elasticsearch features overview
- [LOCAL_SETUP.md](LOCAL_SETUP.md) - Local development setup
- [SETUP.md](SETUP.md) - General configuration guide

## Resources

- [Elastic Cloud Console](https://cloud.elastic.co/)
- [Elastic Cloud Pricing](https://www.elastic.co/pricing/)
- [Elastic Cloud Documentation](https://www.elastic.co/guide/en/cloud/current/index.html)
- [Support Portal](https://support.elastic.co/)
