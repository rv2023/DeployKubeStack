{{/*
Expand the name of the chart.
*/}}
{{- define "deploykubestack.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "deploykubestack.fullname" -}}
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
{{- define "deploykubestack.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "deploykubestack.labels" -}}
helm.sh/chart: {{ include "deploykubestack.chart" . }}
{{ include "deploykubestack.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- with .Values.labels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "deploykubestack.selectorLabels" -}}
app.kubernetes.io/name: {{ include "deploykubestack.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "deploykubestack.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "deploykubestack.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Return the appropriate apiVersion for RBAC APIs
*/}}
{{- define "deploykubestack.rbac.apiVersion" -}}
rbac.authorization.k8s.io/v1
{{- end }}

{{/*
Return the image pull policy
*/}}
{{- define "deploykubestack.imagePullPolicy" -}}
{{- default .Values.image.pullPolicy }}
{{- end }}

{{/*
Return the image tag
*/}}
{{- define "deploykubestack.imageTag" -}}
{{- if .Values.image.tag }}
{{- .Values.image.tag }}
{{- else }}
{{- .Chart.AppVersion }}
{{- end }}
{{- end }}

{{/*
Return the full image with registry
*/}}
{{- define "deploykubestack.image" -}}
{{- printf "%s/%s:%s" .Values.image.registry .Values.image.repository (include "deploykubestack.imageTag" .) }}
{{- end }}
