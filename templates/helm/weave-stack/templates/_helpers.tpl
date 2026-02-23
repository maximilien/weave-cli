{{/*
Expand the name of the chart.
*/}}
{{- define "weave-stack.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "weave-stack.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "weave-stack.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "weave-stack.labels" -}}
helm.sh/chart: {{ include "weave-stack.chart" . }}
{{ include "weave-stack.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "weave-stack.selectorLabels" -}}
app.kubernetes.io/name: {{ include "weave-stack.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
VectorDB labels
*/}}
{{- define "weave-stack.vectordb.labels" -}}
{{ include "weave-stack.labels" . }}
app.kubernetes.io/component: vectordb
{{- end }}

{{/*
VectorDB selector labels
*/}}
{{- define "weave-stack.vectordb.selectorLabels" -}}
{{ include "weave-stack.selectorLabels" . }}
app.kubernetes.io/component: vectordb
app: {{ .Values.vectordb.type }}
{{- end }}
