{{/*
Expand the name of the chart.
*/}}
{{- define "kubeshark.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "kubeshark.fullname" -}}
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
{{- define "kubeshark.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "kubeshark.labels" -}}
helm.sh/chart: {{ include "kubeshark.chart" . }}
{{ include "kubeshark.selectorLabels" . }}
app.kubernetes.io/version: {{ .Chart.Version | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- with .Values.additionalLabels }}
{{ toYaml . }}
{{- end }}
{{- if .Values.tap.labels }}
{{ toYaml .Values.tap.labels }}
{{- end }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "kubeshark.selectorLabels" -}}
app.kubernetes.io/name: {{ include "kubeshark.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "kubeshark.serviceAccountName" -}}
{{- if and .Values.serviceAccount .Values.serviceAccount.create }}
{{- default (include "kubeshark.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- printf "%s-service-account" .Release.Name }}
{{- end }}
{{- end }}
