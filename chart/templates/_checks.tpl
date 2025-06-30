{{- define "checks" -}}
{{- $kubeVersion := lookup "v1" "Namespace" "" "kube-system" }}
{{- if $kubeVersion }}
{{- $certCRD := lookup "apiextensions.k8s.io/v1" "CustomResourceDefinition" "" "certificates.cert-manager.io" }}
{{- if not $certCRD }}
{{- fail "Required CRD 'certificates.cert-manager.io' not found in the cluster. Please install cert-manager first." }}
{{- end }}

{{- $issuerCRD := lookup "apiextensions.k8s.io/v1" "CustomResourceDefinition" "" "issuers.cert-manager.io" }}
{{- if not $issuerCRD }}
{{- fail "Required CRD 'issuers.cert-manager.io' not found in the cluster. Please install cert-manager first." }}
{{- end }}
{{- end }}
{{- end -}}