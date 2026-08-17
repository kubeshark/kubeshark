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

{{/*
Set configmap and secret names based on gitops.enabled
*/}}
{{- define "kubeshark.configmapName" -}}
kubeshark-config-map{{ if .Values.tap.gitops.enabled }}-default{{ end }}
{{- end -}}

{{- define "kubeshark.secretName" -}}
kubeshark-secret{{ if .Values.tap.gitops.enabled }}-default{{ end }}
{{- end -}}


{{/*
Escape double quotes in a string
*/}}
{{- define "kubeshark.escapeDoubleQuotes" -}}
  {{- regexReplaceAll "\"" . "\"" -}}
{{- end -}}

{{/*
Define debug docker tag suffix
*/}}
{{- define "kubeshark.dockerTagDebugVersion" -}}
{{- .Values.tap.pprof.enabled | ternary "-debug" "" }}
{{- end -}}

{{/*
Create docker tag default version
*/}}
{{- define "kubeshark.defaultVersion" -}}
{{- $defaultVersion := (printf "v%s" .Chart.Version) -}}
{{- if .Values.tap.docker.tagLocked }}
  {{- $defaultVersion = regexReplaceAll "^([^.]+\\.[^.]+).*" $defaultVersion "$1" -}}
{{- end }}
{{- $defaultVersion }}
{{- end -}}

{{/*
Set sentry based on internet connectivity and telemetry
*/}}
{{- define "sentry.enabled" -}}
  {{- $sentryEnabledVal := .Values.tap.sentry.enabled -}}
  {{- if not .Values.internetConnectivity -}}
    {{- $sentryEnabledVal = false -}}
  {{- else if not .Values.tap.telemetry.enabled -}}
    {{- $sentryEnabledVal = false -}}
  {{- end -}}
  {{- $sentryEnabledVal -}}
{{- end -}}

{{/*
Dex IdP: retrieve a secret for static client with a specific ID
*/}}
{{- define "getDexKubesharkStaticClientSecret" -}}
  {{- $clientId := .clientId -}}
  {{- range .clients }}
    {{- if eq .id $clientId }}
      {{- .secret }}
    {{- end }}
  {{- end }}
{{- end }}

{{/*
Whether the Hub enforces authentication and authorization on its API.
This is `tap.auth.enabled` and nothing else: licensing, demo mode and the
choice of identity provider do not affect whether the API is gated.
*/}}
{{- define "kubeshark.authEnabled" -}}
{{ .Values.tap.auth.enabled }}
{{- end -}}

{{/*
Reject auth settings that cannot work, instead of rendering a Hub that
authenticates nobody.
*/}}
{{- define "kubeshark.validateAuth" -}}
{{- if .Values.tap.auth.enabled -}}
  {{- if and (eq .Values.tap.auth.type "saml") (empty .Values.tap.auth.saml.idpMetadataUrl) -}}
    {{- fail "tap.auth.enabled is true with tap.auth.type=saml but tap.auth.saml.idpMetadataUrl is empty. Set the IdP metadata URL, or pick another tap.auth.type (oidc, dex, descope)." -}}
  {{- end -}}
  {{- if and (or (eq .Values.tap.auth.type "oidc") (eq .Values.tap.auth.type "dex")) (empty (((.Values.tap).auth).oidc).issuer) -}}
    {{- fail "tap.auth.enabled is true with tap.auth.type=oidc but tap.auth.oidc.issuer is empty. Set the OIDC issuer, or pick another tap.auth.type." -}}
  {{- end -}}
{{- end -}}
{{- end -}}
