# Weave CLI Documentation

Complete documentation for the Weave CLI vector database management tool.

## 📚 Quick Navigation

### Core Documentation

- **[User Guide](USER_GUIDE.md)** - Complete feature documentation and usage guide
- **[Changelog](CHANGELOG.md)** - Version history and release notes
- **[VDB Support Matrix](VDB_SUPPORT.md)** - Feature comparison across databases

### 📖 Guides

User guides for specific features and workflows:

- **[AI Agents](guides/WEAVE_CLI_AI.md)** - Natural language query system with GPT-4o
- **[Batch Processing](guides/BATCH_DOCS_CREATION.md)** - Parallel document processing
- **[Vector DB Abstraction](guides/VECTOR_DB_ABSTRACTION.md)** - Multi-database architecture
- **[Demos](guides/DEMO.md)** - Video tutorials and demonstrations

### 🗄️ Database-Specific Documentation

#### Supabase (PostgreSQL + pgvector)

Complete documentation for Supabase integration:

- **[Supabase Overview](supabase/README.md)** - Getting started and features
- **[Testing Guide](supabase/TESTING.md)** - Integration test setup and usage
- **[Collection Name Fix](supabase/NAME_FIX.md)** - Name preservation implementation
- **[BM25 Improvement](supabase/BM25_IMPROVEMENT.md)** - Full-text search optimization
- **[TODO](supabase/TODO.md)** - Roadmap and planned improvements

#### Weaviate

- **[Integration Status](weaviate/INTEGRATION_STATUS.md)** - Weaviate support details

### 🧪 Testing Documentation

Test coverage and testing guides:

- **[Embedding Test Coverage](tests/EMBEDDING_COVERAGE.md)** - Multi-provider embedding test analysis

### 📝 Planning & Future Work

Analysis and plans for upcoming features:

- **[REPL Progress Improvement](planning/REPL_PROGRESS.md)** - Real-time output streaming for script execution

### 📋 Release Notes

Version-specific release information:

- **[Release Checklist](releases/RELEASE_CHECKLIST.md)** - Release process
- **[v0.3.10](releases/RELEASE_v0.3.10.md)** - Latest release
- **[v0.3.9](releases/RELEASE_v0.3.9.md)**
- **[v0.2.8](releases/RELEASE_v0.2.8.md)**
- **[v0.2.7](releases/RELEASE_v0.2.7.md)**

### 📦 Archive

Historical and reference documentation:

- **[Blog Draft](archive/BLOG_DRAFT.md)** - Technical blog post draft
- **[Presentation](archive/PRESENTATION.md)** - Project presentation
- **[Session Summary](archive/SESSION_SUMMARY.md)** - Development session notes
- **[Weave vs RagMe](archive/WEAVE_VS_RAGME.md)** - Feature comparison
- **[Error Messages](archive/ERROR_MESSAGES.md)** - Error handling reference
- **[Image Metadata Enhancement](archive/IMAGE_METADATA_ENHANCEMENT_PLAN.md)** - Enhancement plan

## 🚀 Getting Started

1. **Installation**: See [User Guide - Installation](USER_GUIDE.md#installation)
2. **Configuration**: See [User Guide - Configuration](USER_GUIDE.md#configuration)
3. **Choose Your Database**: See [VDB Support Matrix](VDB_SUPPORT.md)
4. **Start Using**: Follow [Demos](guides/DEMO.md) or [User Guide](USER_GUIDE.md)

## 🔍 Find What You Need

### By Topic

- **Configuration**: [User Guide - Configuration](USER_GUIDE.md#configuration)
- **AI/Natural Language**: [AI Agents Guide](guides/WEAVE_CLI_AI.md)
- **Batch Processing**: [Batch Docs Guide](guides/BATCH_DOCS_CREATION.md)
- **Testing**: [Supabase Testing](supabase/TESTING.md)
- **Architecture**: [Vector DB Abstraction](guides/VECTOR_DB_ABSTRACTION.md)

### By Database

- **Supabase**: [supabase/](supabase/)
- **Weaviate**: [weaviate/](weaviate/)
- **Mock Database**: [VDB Support Matrix](VDB_SUPPORT.md#mock-database)

### By Use Case

- **Getting Started**: [User Guide](USER_GUIDE.md)
- **Adding Documents**: [User Guide - Documents](USER_GUIDE.md#documents)
- **Searching**: [User Guide - Search](USER_GUIDE.md#search)
- **Managing Collections**: [User Guide - Collections](USER_GUIDE.md#collections)
- **Development**: [Vector DB Abstraction](guides/VECTOR_DB_ABSTRACTION.md)

## 🤝 Contributing

When contributing documentation:

1. Keep the main [User Guide](USER_GUIDE.md) comprehensive but concise
2. Database-specific details go in `supabase/` or `weaviate/` directories
3. Feature guides go in `guides/` directory
4. Update this index when adding new documentation
5. Update internal links when moving files

## 📝 Documentation Structure

```
docs/
├── README.md                    # This file - documentation index
├── USER_GUIDE.md                # Main user guide
├── CHANGELOG.md                 # Version history
├── VDB_SUPPORT.md               # Database feature matrix
│
├── guides/                      # Feature guides
│   ├── WEAVE_CLI_AI.md
│   ├── BATCH_DOCS_CREATION.md
│   ├── VECTOR_DB_ABSTRACTION.md
│   └── DEMO.md
│
├── supabase/                    # Supabase-specific docs
│   ├── README.md
│   ├── TESTING.md
│   ├── NAME_FIX.md
│   ├── BM25_IMPROVEMENT.md
│   └── TODO.md
│
├── weaviate/                    # Weaviate-specific docs
│   └── INTEGRATION_STATUS.md
│
├── tests/                       # Testing documentation
│   └── EMBEDDING_COVERAGE.md
│
├── planning/                    # Future work & analysis
│   └── REPL_PROGRESS.md
│
├── releases/                    # Release notes
│   ├── RELEASE_CHECKLIST.md
│   └── RELEASE_v*.md
│
└── archive/                     # Historical/reference docs
    ├── BLOG_DRAFT.md
    ├── PRESENTATION.md
    └── ...
```

## 📞 Support

- **Issues**: <https://github.com/maximilien/weave-cli/issues>
- **Discussions**: <https://github.com/maximilien/weave-cli/discussions>
- **Main Repository**: <https://github.com/maximilien/weave-cli>
