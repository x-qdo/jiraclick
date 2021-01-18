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
- name: CONFIG_PATH
  value: "/"
- name: HTTPHANDLER_PORT
  value: "8080"
{{- range $key, $value := .Values.clickup }}
- name: CLICKUP_{{$key}}_TOKEN
  valueFrom:
    secretKeyRef:
      key: clickup_{{$key}}_token
      name: {{ $name }}
- name: CLICKUP_{{$key}}_WEBHOOKSECRET
  valueFrom:
    secretKeyRef:
      key: clickup_{{$key}}_webhooksecret
      name: {{ $name }}
{{- end }}
{{- range $key, $value := .Values.jira }}
- name: JIRA_{{$key}}_USERNAME
  valueFrom:
    secretKeyRef:
      key: jira_{{$key}}_username
      name: {{ $name }}
- name: JIRA_{{$key}}_APITOKEN
  valueFrom:
    secretKeyRef:
      key: jira_{{$key}}_apitoken
      name: {{ $name }}
{{- end }}
{{- end -}}

{{- define "jiraclick.config_mount" -}}
- mountPath: "/etc/jiraclick/"
  name: config
  readOnly: true
{{- end -}}

{{- define "jiraclick.config_volume" -}}
- name: config
  secret:
    secretName: {{ include "jiraclick.fullname" . }}
    items:
    - key: config_content
      path: config.yaml
{{- end -}}