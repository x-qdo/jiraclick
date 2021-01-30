{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "jiraclick.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "jiraclick.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "jiraclick.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "jiraclick.labels" -}}
helm.sh/chart: {{ include "jiraclick.chart" . }}
{{ include "jiraclick.selectorLabels" . }}
{{- if .Values.image.tag }}
app.kubernetes.io/version: {{ .Values.image.tag | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "jiraclick.selectorLabels" -}}
app.kubernetes.io/name: {{ include "jiraclick.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "jiraclick.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
    {{ default (include "jiraclick.fullname" .) .Values.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccount.name }}
{{- end -}}
{{- end -}}

{{- define "jiraclick.common_env" -}}
{{- $name := include "jiraclick.fullname" . -}}
- name: DEBUG
  value: "true"
- name: RABBITMQ_URL
  valueFrom:
    secretKeyRef:
      key: rabbitmq_url
      name: {{ $name }}
- name: POSTGRES_URL
  valueFrom:
    secretKeyRef:
      key: postgres_url
      name: {{ $name }}
- name: POSTGRES_INSECURE
  value: "true"
- name: CONFIG_PATH
  value: "/"
- name: HTTPHANDLER_PORT
  value: "8080"
{{- end -}}
