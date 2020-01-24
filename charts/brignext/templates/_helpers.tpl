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

{{- define "brignext.fluentd.fullname" -}}
{{ include "brignext.fullname" . | printf "%s-fluentd" }}
{{- end -}}

{{- define "brignext.apiserver.fullname" -}}
{{ include "brignext.fullname" . | printf "%s-apiserver" }}
{{- end -}}

{{- define "brignext.controller.fullname" -}}
{{ include "brignext.fullname" . | printf "%s-controller" }}
{{- end -}}

{{- define "brignext.fluentd.linux.fullname" -}}
{{ include "brignext.fluentd.fullname" . | printf "%s-linux" }}
{{- end -}}

{{- define "brignext.fluentd.windows.fullname" -}}
{{ include "brignext.fluentd.fullname" . | printf "%s-windows" }}
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

{{- define "brignext.controller.labels" -}}
app.kubernetes.io/component: controller
{{- end -}}

{{- define "brignext.fluentd.labels" -}}
app.kubernetes.io/component: fluentd
{{- end -}}

{{- define "brignext.fluentd.linux.labels" -}}
{{ include "brignext.fluentd.labels" . }}
app.kubernetes.io/os: linux
{{- end -}}

{{- define "brignext.fluentd.windows.labels" -}}
{{ include "brignext.fluentd.labels" . }}
app.kubernetes.io/os: windows
{{- end -}}

{{- define "call-nested" }}
{{- $dot := index . 0 }}
{{- $subchart := index . 1 }}
{{- $template := index . 2 }}
{{- include $template (dict "Chart" (dict "Name" $subchart) "Values" (index $dot.Values $subchart) "Release" $dot.Release "Capabilities" $dot.Capabilities) }}
{{- end }}
