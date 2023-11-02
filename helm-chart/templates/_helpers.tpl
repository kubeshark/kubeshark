{{/*
Expand the name of the chart.
*/}}
{{- define "kubeshark.name" -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "kubeshark.fullname" -}}
{{- printf "%s-%s" .Release.Name .Chart.Name | trunc 63 | trimSuffix "-" }}
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
{{- printf "%s-service-account" .Release.Name }}
{{- end }}
