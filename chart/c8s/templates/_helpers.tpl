{{/*
Expand the name of the chart.
*/}}
{{- define "c8s.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "c8s.fullname" -}}
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
{{- define "c8s.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "c8s.labels" -}}
helm.sh/chart: {{ include "c8s.chart" . }}
{{ include "c8s.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "c8s.selectorLabels" -}}
app.kubernetes.io/name: {{ include "c8s.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "c8s.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "c8s.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Get the namespace
*/}}
{{- define "c8s.namespace" -}}
{{- default .Release.Namespace .Values.global.namespace }}
{{- end }}

{{/*
Get component selector labels
*/}}
{{- define "c8s.componentLabels" -}}
{{ include "c8s.selectorLabels" . }}
app.kubernetes.io/component: {{ .component }}
{{- end }}

{{/*
Create image name
*/}}
{{- define "c8s.image" -}}
{{- $registry := .Values.images.registry }}
{{- $repository := .image.repository }}
{{- $tag := .image.tag }}
{{- printf "%s/%s:%s" $registry $repository $tag }}
{{- end }}

{{/*
Create image pull policy
*/}}
{{- define "c8s.imagePullPolicy" -}}
{{- default .Values.images.pullPolicy .imagePullPolicy }}
{{- end }}
