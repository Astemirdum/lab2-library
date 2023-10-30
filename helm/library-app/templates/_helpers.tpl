{{/*
Expand the name of the chart.
*/}}
{{/*{{- define "library-app.name" -}}*/}}
{{/*{{- default .Chart.Name | trunc 63 | trimSuffix "-" }}*/}}
{{/*{{- end }}*/}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}


{{- define "gateway.fullname" -}}
{{- if .Values.gateway.fullname }}
{{- .Values.gateway.fullname | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.gateway.name }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{- define "library.fullname" -}}
{{- if .Values.library.fullname }}
{{- .Values.library.fullname | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.library.name }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}


{{- define "rating.fullname" -}}
{{- if .Values.rating.fullname }}
{{- .Values.rating.fullname | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.rating.name }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{- define "reservation.fullname" -}}
{{- if .Values.reservation.fullname }}
{{- .Values.reservation.fullname | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.reservation.name }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}



{{- define "db.name" -}}
{{- if .Values.db.name }}
{{- .Values.db.name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.app.name }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-"  }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "library-app.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}



{{/*
Common labels
*/}}
{{- define "library-app.labels" -}}
helm.sh/chart: {{ include "library-app.chart" . }}
{{ include "library-app.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{/*{{- if .Values.version }}*/}}
{{/*app.kubernetes.io/version: {{ .Values.version | quote }}*/}}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}


{{- define "db.labels" -}}
helm.sh/chart: {{ include "library-app.chart" . }}
{{ include "db.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Expand the namespace.
*/}}
{{- define "library-app.namespace" -}}
{{- default .Values.namespace .Release.Namespace | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "db.namespace" -}}
{{- default .Values.db.namespace .Release.Namespace | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "library-app.selectorLabels" -}}
app.kubernetes.io/name: {{ include "gateway.fullname" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}


{{- define "db.selectorLabels" -}}
app.kubernetes.io/name: {{ include "db.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}


{{/*app strategy*/}}
{{- define "library-app.strategy" -}}
rollingUpdate:
  maxSurge: {{ .Values.app.strategy.rollingUpdate.maxSurge}}
  maxUnavailable: {{ .Values.app.strategy.rollingUpdate.maxUnavailable}}
type: {{ .Values.app.strategy.type}}
{{- end }}


{{/*health*/}}
{{- define "library.health" -}}
readinessProbe:
  httpGet: &health
    path: /health
    port: {{ .Values.configData.library.http.port }}
    scheme: HTTP
  initialDelaySeconds: 20
  failureThreshold: 3
  periodSeconds: 30
  timeoutSeconds: 5
livenessProbe:
  httpGet: *health
  failureThreshold: 5
  periodSeconds: 60
  timeoutSeconds: 5
  successThreshold: 1
  initialDelaySeconds: 10
startupProbe:
  failureThreshold: 10
  httpGet: *health
  periodSeconds: 10
  timeoutSeconds: 5
  successThreshold: 1
{{- end }}

{{/*health*/}}
{{- define "gateway.health" -}}
readinessProbe:
  httpGet: &health
    path: /health
    port: {{ .Values.configData.gateway.http.port }}
    scheme: HTTP
  initialDelaySeconds: 20
  failureThreshold: 3
  periodSeconds: 30
  timeoutSeconds: 5
livenessProbe:
  httpGet: *health
  failureThreshold: 5
  periodSeconds: 60
  timeoutSeconds: 5
  successThreshold: 1
  initialDelaySeconds: 10
startupProbe:
  failureThreshold: 10
  httpGet: *health
  periodSeconds: 10
  timeoutSeconds: 5
  successThreshold: 1
{{- end }}


{{/*health*/}}
{{- define "rating.health" -}}
readinessProbe:
  httpGet: &health
    path: /health
    port: {{ .Values.configData.rating.http.port }}
    scheme: HTTP
  initialDelaySeconds: 20
  failureThreshold: 3
  periodSeconds: 30
  timeoutSeconds: 5
livenessProbe:
  httpGet: *health
  failureThreshold: 5
  periodSeconds: 60
  timeoutSeconds: 5
  successThreshold: 1
  initialDelaySeconds: 10
startupProbe:
  failureThreshold: 10
  httpGet: *health
  periodSeconds: 10
  timeoutSeconds: 5
  successThreshold: 1
{{- end }}

{{/*health*/}}
{{- define "reservation.health" -}}
readinessProbe:
  httpGet: &health
    path: /health
    port: {{ .Values.configData.reservation.http.port }}
    scheme: HTTP
  initialDelaySeconds: 20
  failureThreshold: 3
  periodSeconds: 30
  timeoutSeconds: 5
livenessProbe:
  httpGet: *health
  failureThreshold: 5
  periodSeconds: 60
  timeoutSeconds: 5
  successThreshold: 1
  initialDelaySeconds: 10
startupProbe:
  failureThreshold: 10
  httpGet: *health
  periodSeconds: 10
  timeoutSeconds: 5
  successThreshold: 1
{{- end }}


{{- define "db.health" -}}
livenessProbe:
  exec:
    command:
      - bash
      - -ec
      - 'PGPASSWORD=$POSTGRES_PASSWORD psql -w -U "${POSTGRES_USER}" -d "${POSTGRES_DB}"  -h 127.0.0.1 -c "SELECT 1"'
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5
  successThreshold: 1
  failureThreshold: 6
readinessProbe:
  exec:
    command:
      - bash
      - -ec
      - 'PGPASSWORD=$POSTGRES_PASSWORD psql -w -U "${POSTGRES_USER}" -d "${POSTGRES_DB}"  -h 127.0.0.1 -c "SELECT 1"'
  initialDelaySeconds: 5
  periodSeconds: 10
  timeoutSeconds: 5
  successThreshold: 1
  failureThreshold: 6
{{- end}}


{{/*
env pgHostPortDB
*/}}
{{- define "app.env.pgHostPortDB" -}}
- name: DB_HOST
  valueFrom:
    configMapKeyRef:
      name: {{ include "gateway.fullname" . }}-config
      key: DB_HOST
- name: DB_PORT
  valueFrom:
    configMapKeyRef:
      name: {{ include "gateway.fullname" . }}-config
      key: DB_PORT
{{- end }}



{{/*
Wait for DB to be ready
*/}}
{{- define "app.pgWait" -}}
['sh', '-c', "until nc -w 2 $(DB_HOST) $(DB_PORT); do echo Waiting for $(DB_HOST):$(DB_PORT) to be ready; sleep 5; done"]
{{- end }}
