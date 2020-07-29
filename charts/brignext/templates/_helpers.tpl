{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "brignext.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "brignext.fullname" -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{- define "brignext.logger.fullname" -}}
{{ include "brignext.fullname" . | printf "%s-logger" }}
{{- end -}}

{{- define "brignext.apiserver.fullname" -}}
{{ include "brignext.fullname" . | printf "%s-apiserver" }}
{{- end -}}

{{- define "brignext.scheduler.fullname" -}}
{{ include "brignext.fullname" . | printf "%s-scheduler" }}
{{- end -}}

{{- define "brignext.artemis.fullname" -}}
{{ include "brignext.fullname" . | printf "%s-artemis" }}
{{- end -}}

{{- define "brignext.logger.linux.fullname" -}}
{{ include "brignext.logger.fullname" . | printf "%s-linux" }}
{{- end -}}

{{- define "brignext.logger.windows.fullname" -}}
{{ include "brignext.logger.fullname" . | printf "%s-windows" }}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "brignext.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "brignext.labels" -}}
helm.sh/chart: {{ include "brignext.chart" . }}
{{ include "brignext.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "brignext.selectorLabels" -}}
app.kubernetes.io/name: {{ include "brignext.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "brignext.apiserver.labels" -}}
app.kubernetes.io/component: apiserver
{{- end -}}

{{- define "brignext.scheduler.labels" -}}
app.kubernetes.io/component: scheduler
{{- end -}}

{{- define "brignext.artemis.labels" -}}
app.kubernetes.io/component: artemis
{{- end -}}

{{- define "brignext.artemis.primary.labels" -}}
{{ include "brignext.artemis.labels" . }}
app.kubernetes.io/role: primary
{{- end -}}

{{- define "brignext.artemis.secondary.labels" -}}
{{ include "brignext.artemis.labels" . }}
app.kubernetes.io/role: secondary
{{- end -}}

{{- define "brignext.logger.labels" -}}
app.kubernetes.io/component: logger
{{- end -}}

{{- define "brignext.logger.linux.labels" -}}
{{ include "brignext.logger.labels" . }}
app.kubernetes.io/os: linux
{{- end -}}

{{- define "brignext.logger.windows.labels" -}}
{{ include "brignext.logger.labels" . }}
app.kubernetes.io/os: windows
{{- end -}}

{{- define "call-nested" }}
{{- $dot := index . 0 }}
{{- $subchart := index . 1 }}
{{- $template := index . 2 }}
{{- include $template (dict "Chart" (dict "Name" $subchart) "Values" (index $dot.Values $subchart) "Release" $dot.Release "Capabilities" $dot.Capabilities) }}
{{- end }}
