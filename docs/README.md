# Weave CLI Documentation

Complete documentation for the Weave CLI vector database management tool.

## Quick Navigation

### Core Documentation

- **[User Guide](USER_GUIDE.md)** - Complete feature documentation and usage guide
- **[Architecture](ARCHITECTURE.md)** - System design, agent layer, orchestration
- **[Roadmap](ROADMAP.md)** - Version history, upcoming features, and release plan
- **[VDB Support Matrix](VDB_SUPPORT_MATRIX.md)** - Quick reference feature comparison
- **[VDB Support](VDB_SUPPORT.md)** - Detailed database comparison and analysis

### Guides

User guides for specific features and workflows:

- **[Backup & Restore](guides/BACKUP_RESTORE.md)** - Complete backup/restore guide with examples
- **[Weave Stack Quickstart](guides/WEAVE_STACK_QUICKSTART.md)** - Local development stack setup
- **[AI Agents](guides/WEAVE_CLI_AI.md)** - Natural language query system with GPT-4o
- **[Batch Processing](guides/BATCH_DOCS_CREATION.md)** - Parallel document processing
- **[Vector DB Abstraction](guides/VECTOR_DB_ABSTRACTION.md)** - Multi-database architecture
- **[BM25 Support by VDB](guides/BM25_SUPPORT_BY_VDB.md)** - Full-text search across all databases
- **[OSS Embedding Tips](guides/OSS_EMBEDDING_TESTING_TIPS.md)** - Free embedding providers
- **[Demos](guides/DEMO.md)** - Video tutorials and demonstrations

### Database-Specific Documentation

All VDB docs are under [`vdbs/`](vdbs/):

| Database | Docs | Status |
|----------|------|--------|
| **Weaviate** | [Setup](vdbs/weaviate/SETUP.md), [Integration Status](vdbs/weaviate/INTEGRATION_STATUS.md) | Stable |
| **Qdrant** | [Setup](vdbs/qdrant/SETUP.md) | Stable |
| **Milvus** | [Overview](vdbs/milvus/README.md), [Setup](vdbs/milvus/SETUP.md), [Local](vdbs/milvus/LOCAL_SETUP.md), [Cloud](vdbs/milvus/CLOUD_SETUP.md) | Stable |
| **Chroma** | [Setup](vdbs/chroma/SETUP.md) | Stable |
| **Supabase** | [Overview](vdbs/supabase/README.md), [Setup](vdbs/supabase/SETUP.md), [Testing](vdbs/supabase/TESTING.md), [TODO](vdbs/supabase/TODO.md) | Stable |
| **MongoDB** | [Overview](vdbs/mongodb/README.md), [Setup](vdbs/mongodb/SETUP.md), [Atlas](vdbs/mongodb/ATLAS_SETUP.md) | Stable |
| **Neo4j** | [Overview](vdbs/neo4j/README.md), [Setup](vdbs/neo4j/SETUP.md), [Aura](vdbs/neo4j/AURA_SETUP.md) | Stable |
| **Pinecone** | [Setup](vdbs/pinecone/SETUP.md) | Beta |
| **Elasticsearch** | [Overview](vdbs/elasticsearch/README.md), [Setup](vdbs/elasticsearch/SETUP.md), [Local](vdbs/elasticsearch/LOCAL_SETUP.md), [Cloud](vdbs/elasticsearch/CLOUD_SETUP.md) | Beta |
| **OpenSearch** | [Overview](vdbs/opensearch/README.md), [AWS](vdbs/opensearch/AWS_SETUP.md) | Stable |

### Other Documentation

- **[Agent Management](AGENT_MANAGEMENT.md)** - Built-in and custom agents
- **[Weave Stack](WEAVE_STACK.md)** - Kubernetes stack deployment
- **[MCP AI Tools](mcp/MCP_AI_TOOLS.md)** - MCP server integration
- **[Observability](OBSERVABILITY.md)** - Opik tracing and monitoring
- **[Shell Completion](SHELL_COMPLETION.md)** - Tab completion setup
- **[Timeout Configuration](TIMEOUT_CONFIGURATION.md)** - Per-operation timeout tuning
- **[VDB Naming Convention](VDB_NAMING_CONVENTION.md)** - Collection naming rules
- **[Production Ready](PRODUCTION_READY.md)** - Production readiness checklist
- **[Client0 Getting Started](CLIENT0_GETTING_STARTED.md)** - Client0 onboarding

### Testing

- **[Test Guide](TEST_GUIDE.md)** - Running tests
- **[Embedding Coverage](tests/EMBEDDING_COVERAGE.md)** - Multi-provider embedding test analysis

### Integrations

- **[GitHub Actions](integrations/GITHUB_ACTIONS.md)**
- **[Airflow](integrations/AIRFLOW.md)**
- **[Argo Workflows](integrations/ARGO_WORKFLOWS.md)**

### Release Notes

- **[Release Checklist](releases/RELEASE_CHECKLIST.md)** - Release process
- **[v0.11.5](releases/v0.11.5-RELEASE-NOTES.md)** | [v0.11.2](releases/RELEASE_v0.11.2.md) | [v0.11.1](releases/RELEASE_v0.11.1.md) | [v0.11.0](releases/RELEASE_v0.11.0.md)
- **[v0.9.15](releases/RELEASE_v0.9.15.md)** | [v0.9.13](releases/RELEASE_v0.9.13.md) | [v0.9.1](releases/RELEASE_v0.9.1.md)
- **[v0.7.2](releases/RELEASE_v0.7.2.md)** | [v0.7.1](releases/RELEASE_v0.7.1.md) | [v0.7.0](releases/RELEASE_v0.7.0.md)
- Older releases: [archive/releases/](archive/releases/)

### Collaboration

- **[Opik Video/Blog Checklist](blogs/OPIK_VIDEO_BLOG_CHECKLIST.md)** - Opik collaboration (deadline: Mar 24-31)
- **[Presentation](PRESENTATION.md)** - Project presentation slides
- **[Blog Draft](archive/BLOG_DRAFT.md)** - Technical blog post draft

### Planning

Active planning docs:

- **[Planning Index](planning/README.md)**
- **[Weave Stack Phase 2](planning/WEAVE_STACK_PHASE_2_PLAN.md)**
- **[Multi-VDB Support Plan](planning/MULTI_VDB_SUPPORT_PLAN.md)**

### Archive

Historical documentation: **[archive/](archive/)**

## Documentation Structure

```
docs/
├── README.md                    # This file - documentation index
├── USER_GUIDE.md                # Main user guide
├── ARCHITECTURE.md              # System architecture
├── VDB_SUPPORT.md               # Detailed database comparison
├── VDB_SUPPORT_MATRIX.md        # Quick reference matrix
├── WEAVE_STACK.md               # Stack deployment
│
├── vdbs/                        # All VDB-specific docs
│   ├── weaviate/
│   ├── qdrant/
│   ├── milvus/
│   ├── chroma/
│   ├── supabase/
│   ├── mongodb/
│   ├── neo4j/
│   ├── pinecone/
│   ├── elasticsearch/
│   └── opensearch/
│
├── guides/                      # Feature guides
├── blogs/                       # Blog/video collaboration
├── examples/                    # Usage examples
├── integrations/                # CI/CD integrations
├── tests/                       # Testing documentation
├── planning/                    # Active planning
├── releases/                    # Release notes
├── mcp/                         # MCP integration
└── archive/                     # Historical docs
```

## Support

- **Issues**: <https://github.com/maximilien/weave-cli/issues>
- **Discussions**: <https://github.com/maximilien/weave-cli/discussions>
- **Main Repository**: <https://github.com/maximilien/weave-cli>
