// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package stack

import "time"

// StackConfig represents the complete weave-stack.yaml configuration
type StackConfig struct {
	Version        string               `yaml:"version"`
	Name           string               `yaml:"name"`
	Description    string               `yaml:"description,omitempty"`
	Runtime        RuntimeConfig        `yaml:"runtime"`
	Infrastructure InfrastructureConfig `yaml:"infrastructure"`
	Collections    []CollectionConfig   `yaml:"collections,omitempty"`
	Ingestion      *IngestionConfig     `yaml:"ingestion,omitempty"`
	Dashboard      *DashboardConfig     `yaml:"dashboard,omitempty"`
	Operations     *OperationsConfig    `yaml:"operations,omitempty"`
}

// RuntimeConfig defines where the stack will run
type RuntimeConfig struct {
	Kubernetes       KubernetesConfig `yaml:"kubernetes"`
	ContainerRuntime string           `yaml:"container_runtime"` // "podman" or "docker"
}

// KubernetesConfig defines K8s cluster configuration
type KubernetesConfig struct {
	Provider string          `yaml:"provider"` // "kind", "minikube", "eks", "gke"
	Kind     *KindConfig     `yaml:"kind,omitempty"`
	Minikube *MinikubeConfig `yaml:"minikube,omitempty"`
	EKS      *EKSConfig      `yaml:"eks,omitempty"`
	GKE      *GKEConfig      `yaml:"gke,omitempty"`
}

// KindConfig is Kind-specific configuration
type KindConfig struct {
	Name   string `yaml:"name"`
	Nodes  int    `yaml:"nodes"`
	Config string `yaml:"config,omitempty"` // Custom Kind config YAML
}

// MinikubeConfig is Minikube-specific configuration
type MinikubeConfig struct {
	Driver string   `yaml:"driver"` // "podman", "docker", "kvm2", "hyperkit"
	CPUs   int      `yaml:"cpus"`
	Memory string   `yaml:"memory"` // e.g., "16384" (MB)
	Addons []string `yaml:"addons,omitempty"`
}

// EKSConfig is AWS EKS configuration
type EKSConfig struct {
	Region     string         `yaml:"region"`
	NodeGroups []EKSNodeGroup `yaml:"node_groups"`
}

// EKSNodeGroup defines an EKS node group
type EKSNodeGroup struct {
	Name         string `yaml:"name"`
	InstanceType string `yaml:"instance_type"`
	DesiredSize  int    `yaml:"desired_size"`
	MinSize      int    `yaml:"min_size"`
	MaxSize      int    `yaml:"max_size"`
}

// GKEConfig is GCP GKE configuration
type GKEConfig struct {
	Zone      string        `yaml:"zone"`
	NodePools []GKENodePool `yaml:"node_pools"`
}

// GKENodePool defines a GKE node pool
type GKENodePool struct {
	Name        string       `yaml:"name"`
	MachineType string       `yaml:"machine_type"`
	NodeCount   int          `yaml:"node_count"`
	AutoScaling *AutoScaling `yaml:"auto_scaling,omitempty"`
}

// AutoScaling defines auto-scaling configuration
type AutoScaling struct {
	MinNodeCount int `yaml:"min_node_count"`
	MaxNodeCount int `yaml:"max_node_count"`
}

// InfrastructureConfig defines infrastructure components
type InfrastructureConfig struct {
	VectorDB     VectorDBConfig      `yaml:"vectordb"`
	LLM          *LLMConfig          `yaml:"llm,omitempty"`
	ImageStorage *ImageStorageConfig `yaml:"image_storage,omitempty"`
}

// VectorDBConfig defines vector database configuration
type VectorDBConfig struct {
	Type       string                 `yaml:"type"` // "milvus", "qdrant", "weaviate"
	Version    string                 `yaml:"version"`
	Kubernetes *KubernetesResources   `yaml:"kubernetes,omitempty"`
	Resources  ResourceRequirements   `yaml:"resources"`
	Config     map[string]interface{} `yaml:"config,omitempty"`
	Health     *HealthCheck           `yaml:"health,omitempty"`
}

// KubernetesResources defines K8s-specific resource configuration
type KubernetesResources struct {
	Deployment DeploymentConfig `yaml:"deployment"`
}

// DeploymentConfig defines K8s Deployment settings
type DeploymentConfig struct {
	Replicas     int    `yaml:"replicas"`
	StorageClass string `yaml:"storage_class"`
	PVCSize      string `yaml:"pvc_size"`
}

// ResourceRequirements defines CPU and memory requirements
type ResourceRequirements struct {
	Requests ResourceLimits `yaml:"requests"`
	Limits   ResourceLimits `yaml:"limits"`
}

// ResourceLimits defines resource limits
type ResourceLimits struct {
	Memory string `yaml:"memory"`
	CPU    string `yaml:"cpu"`
}

// HealthCheck defines health check configuration
type HealthCheck struct {
	Liveness  *Probe `yaml:"liveness,omitempty"`
	Readiness *Probe `yaml:"readiness,omitempty"`
}

// Probe defines a K8s probe
type Probe struct {
	HTTPGet             *HTTPGetAction `yaml:"http_get,omitempty"`
	InitialDelaySeconds int            `yaml:"initial_delay_seconds,omitempty"`
	PeriodSeconds       int            `yaml:"period_seconds,omitempty"`
	TimeoutSeconds      int            `yaml:"timeout_seconds,omitempty"`
	FailureThreshold    int            `yaml:"failure_threshold,omitempty"`
}

// HTTPGetAction defines an HTTP GET probe
type HTTPGetAction struct {
	Path string `yaml:"path"`
	Port int    `yaml:"port"`
}

// LLMConfig defines LLM provider configuration
type LLMConfig struct {
	Provider string                 `yaml:"provider"` // "openai", "anthropic", "ollama"
	Models   LLMModels              `yaml:"models"`
	Config   map[string]interface{} `yaml:"config,omitempty"`
}

// LLMModels defines LLM model names
type LLMModels struct {
	Embedding string `yaml:"embedding"`
	Chat      string `yaml:"chat"`
}

// ImageStorageConfig defines image storage configuration
type ImageStorageConfig struct {
	Type      string                `yaml:"type"` // "minio", "s3", "local"
	Bucket    string                `yaml:"bucket,omitempty"`
	Endpoint  string                `yaml:"endpoint,omitempty"`
	Resources *ResourceRequirements `yaml:"resources,omitempty"`
}

// CollectionConfig defines a data collection
type CollectionConfig struct {
	Name        string           `yaml:"name"`
	Type        string           `yaml:"type"` // "text", "image"
	Description string           `yaml:"description,omitempty"`
	Schema      SchemaConfig     `yaml:"schema"`
	Sources     []DataSource     `yaml:"sources"`
	Chunking    *ChunkingConfig  `yaml:"chunking,omitempty"`
	Embedding   *EmbeddingConfig `yaml:"embedding,omitempty"`
}

// SchemaConfig defines collection schema
type SchemaConfig struct {
	VectorDimensions int           `yaml:"vector_dimensions"`
	Fields           []SchemaField `yaml:"fields,omitempty"`
}

// SchemaField defines a schema field
type SchemaField struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"` // "string", "int", "float", "text", "bool"
}

// DataSource defines where data comes from
type DataSource struct {
	Pattern string                 `yaml:"pattern"`        // Glob pattern
	Type    string                 `yaml:"type"`           // "pdf", "docx", "txt"
	Mode    string                 `yaml:"mode,omitempty"` // "text-only", "image-extraction"
	Options map[string]interface{} `yaml:"options,omitempty"`
}

// ChunkingConfig defines how to chunk documents
type ChunkingConfig struct {
	Strategy     string `yaml:"strategy"` // "semantic", "fixed", "sentence"
	ChunkSize    int    `yaml:"chunk_size"`
	ChunkOverlap int    `yaml:"chunk_overlap,omitempty"`
}

// EmbeddingConfig defines embedding model
type EmbeddingConfig struct {
	Model    string `yaml:"model"`
	Provider string `yaml:"provider"`
}

// IngestionConfig defines data ingestion pipeline
type IngestionConfig struct {
	ParallelWorkers int               `yaml:"parallel_workers,omitempty"`
	BatchSize       int               `yaml:"batch_size,omitempty"`
	Retry           *RetryConfig      `yaml:"retry,omitempty"`
	Timeout         *TimeoutConfig    `yaml:"timeout,omitempty"`
	Checkpoint      *CheckpointConfig `yaml:"checkpoint,omitempty"`
	Monitoring      *MonitoringConfig `yaml:"monitoring,omitempty"`
	Phases          []PhaseConfig     `yaml:"phases,omitempty"`
}

// RetryConfig defines retry behavior
type RetryConfig struct {
	MaxAttempts int           `yaml:"max_attempts"`
	Backoff     string        `yaml:"backoff"` // "constant", "exponential"
	BaseDelay   time.Duration `yaml:"base_delay,omitempty"`
}

// TimeoutConfig defines timeout limits
type TimeoutConfig struct {
	PerFile time.Duration `yaml:"per_file,omitempty"`
	Total   time.Duration `yaml:"total,omitempty"`
}

// CheckpointConfig defines checkpointing
type CheckpointConfig struct {
	Enabled      bool   `yaml:"enabled"`
	Path         string `yaml:"path,omitempty"`
	SaveInterval int    `yaml:"save_interval,omitempty"`
}

// MonitoringConfig defines monitoring settings
type MonitoringConfig struct {
	VDBHealthCheck  bool   `yaml:"vdb_health_check,omitempty"`
	RestartOnOOM    bool   `yaml:"restart_on_oom,omitempty"`
	MemoryThreshold string `yaml:"memory_threshold,omitempty"`
}

// PhaseConfig defines an ingestion phase
type PhaseConfig struct {
	Name                   string   `yaml:"name"`
	Collections            []string `yaml:"collections"`
	Parallel               bool     `yaml:"parallel,omitempty"`
	RestartVDBBetweenFiles bool     `yaml:"restart_vdb_between_files,omitempty"`
}

// DashboardConfig defines dashboard configuration
type DashboardConfig struct {
	Enabled bool          `yaml:"enabled"`
	Type    string        `yaml:"type"` // "web", "cli", "none"
	Web     *WebDashboard `yaml:"web,omitempty"`
}

// WebDashboard defines web dashboard configuration
type WebDashboard struct {
	Framework  string               `yaml:"framework"` // "nextjs", "react", "vue"
	Language   string               `yaml:"language"`  // "typescript", "javascript"
	Version    string               `yaml:"version,omitempty"`
	Kubernetes *KubernetesResources `yaml:"kubernetes,omitempty"`
	Features   []string             `yaml:"features,omitempty"`
	API        []APIRoute           `yaml:"api,omitempty"`
	EnvFile    string               `yaml:"env_file,omitempty"`
	EnvVars    map[string]string    `yaml:"env_vars,omitempty"`
	Build      *BuildConfig         `yaml:"build,omitempty"`
	Dev        *DevConfig           `yaml:"dev,omitempty"`
}

// APIRoute defines an API route
type APIRoute struct {
	Path       string   `yaml:"path"`
	Collection string   `yaml:"collection,omitempty"`
	Type       string   `yaml:"type,omitempty"` // "search", "healthcheck"
	Methods    []string `yaml:"methods,omitempty"`
	Auth       string   `yaml:"auth,omitempty"` // "required", "optional", "none"
}

// BuildConfig defines build configuration
type BuildConfig struct {
	Dockerfile string `yaml:"dockerfile,omitempty"`
	Context    string `yaml:"context,omitempty"`
	Target     string `yaml:"target,omitempty"`
	Cache      bool   `yaml:"cache,omitempty"`
}

// DevConfig defines development mode configuration
type DevConfig struct {
	HotReload bool   `yaml:"hot_reload,omitempty"`
	Port      int    `yaml:"port,omitempty"`
	Command   string `yaml:"command,omitempty"`
}

// OperationsConfig defines operational settings
type OperationsConfig struct {
	Logging *LoggingConfig `yaml:"logging,omitempty"`
	Cleanup *CleanupConfig `yaml:"cleanup,omitempty"`
	Backup  *BackupConfig  `yaml:"backup,omitempty"`
}

// LoggingConfig defines logging settings
type LoggingConfig struct {
	Level   string      `yaml:"level"`  // "debug", "info", "warn", "error"
	Format  string      `yaml:"format"` // "text", "json"
	Outputs []LogOutput `yaml:"outputs"`
}

// LogOutput defines a logging output
type LogOutput struct {
	Type string `yaml:"type"` // "file", "stdout"
	Path string `yaml:"path,omitempty"`
}

// CleanupConfig defines cleanup policies
type CleanupConfig struct {
	RemoveCheckpointsOnSuccess bool `yaml:"remove_checkpoints_on_success,omitempty"`
	RetainLogsDays             int  `yaml:"retain_logs_days,omitempty"`
	CompressOldLogs            bool `yaml:"compress_old_logs,omitempty"`
}

// BackupConfig defines backup configuration
type BackupConfig struct {
	Enabled       bool   `yaml:"enabled"`
	Schedule      string `yaml:"schedule,omitempty"`
	RetentionDays int    `yaml:"retention_days,omitempty"`
}

// ClusterInfo stores information about the active cluster
type ClusterInfo struct {
	Name             string    `json:"name"`
	Provider         string    `json:"provider"` // "kind", "minikube", "eks", "gke"
	Context          string    `json:"context"`  // kubectl context name
	ContainerRuntime string    `json:"container_runtime"`
	CreatedAt        time.Time `json:"created_at"`
	Status           string    `json:"status"` // "active", "stopped", "error"
}

// StackState stores the current state of the stack
type StackState struct {
	Cluster   *ClusterInfo             `json:"cluster,omitempty"`
	Services  map[string]ServiceStatus `json:"services"`
	Ingestion *IngestionState          `json:"ingestion,omitempty"`
	Dashboard *DashboardState          `json:"dashboard,omitempty"`
	UpdatedAt time.Time                `json:"updated_at"`
}

// ServiceStatus represents the status of a service
type ServiceStatus struct {
	Name      string `json:"name"`
	Status    string `json:"status"` // "running", "stopped", "error"
	Health    string `json:"health"` // "healthy", "unhealthy", "unknown"
	Uptime    string `json:"uptime,omitempty"`
	Resources string `json:"resources,omitempty"`
}

// IngestionState stores ingestion progress
type IngestionState struct {
	Phase          string       `json:"phase"`
	CompletedFiles []string     `json:"completed_files"`
	FailedFiles    []FailedFile `json:"failed_files,omitempty"`
	CurrentFile    string       `json:"current_file,omitempty"`
	Progress       Progress     `json:"progress"`
	StartedAt      time.Time    `json:"started_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}

// FailedFile represents a failed ingestion attempt
type FailedFile struct {
	File     string `json:"file"`
	Error    string `json:"error"`
	Attempts int    `json:"attempts"`
}

// Progress represents ingestion progress
type Progress struct {
	TotalFiles int `json:"total_files"`
	Completed  int `json:"completed"`
	Failed     int `json:"failed"`
	Remaining  int `json:"remaining"`
}

// DashboardState stores dashboard status
type DashboardState struct {
	Status string `json:"status"` // "running", "stopped", "error"
	URL    string `json:"url,omitempty"`
	Port   int    `json:"port,omitempty"`
}
