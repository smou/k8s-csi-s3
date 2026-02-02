{{- define "log.level"}}
{{- $logLevel := default 0 .Values.verbose -}}
{{- cat "--v=" $logLevel | nospace | quote}}
{{- end}}

{{- define "driver.name"}}
{{- default "minio-csi-s3" .Values.nameOverride | trunc 63 | trimSuffix "-"}}
{{- end}}

{{- define "driver.image"}}
{{- $imagename := default "cshoch3/k8s-csi-s3-minio" -}}
{{- $imageversion := default .Chart.AppVersion .Values.version -}}
{{- cat $imagename ":" $imageversion | nospace}}
{{- end}}

{{- define "driver.configmap.name" -}}
  {{- printf "%s-config" (include "driver.name" .) -}}
{{- end -}}

{{- define "driver.secret.name" -}}
  {{- printf "%s-secret" (include "driver.name" .) -}}
{{- end -}}

{{- define "driver.labels"}}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/part-of: {{ include "driver.name" . }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end}}

{{- define "driver.namespace.name"}}
{{- $name := .Release.Namespace -}}
{{- if .Values.namespace -}}
  {{- $name = default .Release.Namespace .Values.namespace.name -}}
{{- end -}}
{{- $name | trunc 63 | trimSuffix "-" -}}
{{- end}}

{{- define "driver.namespace.create" -}}
    {{- $nsName := include "driver.namespace.name" . -}}
    {{- if and .Values.namespace .Values.namespace.create (ne $nsName "default") -}}
        {{- printf "true" -}}
    {{- else -}}
        {{- printf "false" -}}
    {{- end -}}
{{- end -}}

{{- define "driver.namespace.annotations"}}
meta.helm.sh/release-name: {{ include "driver.name" . | quote}}
meta.helm.sh/release-namespace: {{ include "driver.namespace.name" . | quote}}
{{- end}}

{{- define "driver.namespace.labels"}}
app.kubernetes.io/managed-by: {{ .Release.Service }}
kubernetes.io/metadata.name: {{ include "driver.namespace.name" . | quote}}
{{- end}}

{{- define "driver.controller.name" -}}
  {{- printf "%s-controller" (include "driver.name" .) -}}
{{- end -}}

{{- define "driver.node.name" -}}
  {{- printf "%s-node" (include "driver.name" .) -}}
{{- end -}}

{{- define "driver.serviceaccount.name" -}}
  {{- printf "%s" (include "driver.name" .) -}}
{{- end -}}

{{- define "driver.role.controller.name" -}}
  {{- printf "%s-role" (include "driver.controller.name" .) -}}
{{- end -}}

{{- define "driver.role.node.name" -}}
  {{- printf "%s-role" (include "driver.node.name" .) -}}
{{- end -}}