{{/*
Expand the name of the chart.
*/}}
{{- define "pganalytics.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "pganalytics.fullname" -}}
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
{{- define "pganalytics.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "pganalytics.labels" -}}
helm.sh/chart: {{ include "pganalytics.chart" . }}
{{ include "pganalytics.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- with .Values.commonLabels }}
{{ toYaml . | indent 0 }}
{{- end }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "pganalytics.selectorLabels" -}}
app.kubernetes.io/name: {{ include "pganalytics.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "pganalytics.serviceAccountName" -}}
{{- if .Values.rbac.serviceAccount.create }}
{{- default (include "pganalytics.fullname" .) .Values.rbac.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.rbac.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Return the namespace
*/}}
{{- define "pganalytics.namespace" -}}
{{- default .Release.Namespace .Values.namespace.name }}
{{- end }}

{{/*
Return the appropriate apiVersion for RBAC APIs
*/}}
{{- define "pganalytics.rbac.apiVersion" -}}
{{- if .Capabilities.APIVersions.Has "rbac.authorization.k8s.io/v1" }}
{{- print "rbac.authorization.k8s.io/v1" }}
{{- else }}
{{- print "rbac.authorization.k8s.io/v1beta1" }}
{{- end }}
{{- end }}

{{/*
Return true if we should create RBAC resources
*/}}
{{- define "pganalytics.rbac.create" -}}
{{- if .Values.rbac.create -}}
true
{{- end }}
{{- end }}

{{/*
Return the database host
*/}}
{{- define "pganalytics.database.host" -}}
{{- if .Values.postgresql.enabled }}
{{- print "postgresql." (include "pganalytics.namespace" .) ".svc.cluster.local" }}
{{- else }}
{{- required "PostgreSQL host must be specified if postgresql.enabled is false" .Values.externalPostgresql.host }}
{{- end }}
{{- end }}

{{/*
Return the Redis host
*/}}
{{- define "pganalytics.redis.host" -}}
{{- if .Values.redis.enabled }}
{{- print "redis." (include "pganalytics.namespace" .) ".svc.cluster.local" }}
{{- else }}
{{- required "Redis host must be specified if redis.enabled is false" .Values.externalRedis.host }}
{{- end }}
{{- end }}

{{/*
Return the backend service endpoint
*/}}
{{- define "pganalytics.backend.endpoint" -}}
{{- print "http://" (include "pganalytics.fullname" .) "-backend." (include "pganalytics.namespace" .) ".svc.cluster.local:" (.Values.backend.service.port | default 8080) }}
{{- end }}
