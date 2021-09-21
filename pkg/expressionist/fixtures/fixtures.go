package fixtures

import (
	naisiov1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	validAction      = "kubectl describe pod {{ $labels.kubernetes_pod_name }} -n {{ $labels.kubernetes_namespace }}` for events, og `kubectl logs {{ $labels.kubernetes_pod_name }} -n {{ $labels.kubernetes_namespace }}` for logger"
	validDescription = "App {{ $labels.app }} er nede i namespace {{ $labels.kubernetes_namespace }}"
	validExpr        = "kube_deployment_status_replicas_unavailable{deployment=\"nais-testapp\"} > 0"
)

func EmptySpec() *naisiov1.Alert {
	return &naisiov1.Alert{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Alert",
			APIVersion: "nais.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nais-testapp",
			Namespace: "aura",
			Labels: map[string]string{
				"team": "aura",
			},
		},
		Spec: naisiov1.AlertSpec{},
	}
}

func ConfigurationTemplate(action, description, expr string) *naisiov1.Alert {
	alert := EmptySpec()
	alert.Spec = naisiov1.AlertSpec{
		Route: naisiov1.Route{
			GroupWait:      "30s",
			GroupInterval:  "5m",
			RepeatInterval: "3h",
			GroupBy:        []string{"<label_name>"},
		},
		Receivers: naisiov1.Receivers{
			Slack: naisiov1.Slack{
				Channel:     "#alert-channel",
				PrependText: "Oh noes!",
			},
			Email: naisiov1.Email{
				To: "myteam@nav.no",
			},
			SMS: naisiov1.SMS{
				Recipients: "12345678",
			},
		},
		Alerts: []naisiov1.Rule{
			{
				Alert:         "applikasjon nede",
				Description:   description,
				Expr:          expr,
				For:           "2m",
				Action:        action,
				Documentation: "https://doc.nais.io/observability/alerts/",
				SLA:           "Mellom 8 og 16",
				Severity:      "danger",
				Priority:      "0",
			},
		},
		InhibitRules: []naisiov1.InhibitRules{
			{
				Targets: map[string]string{
					"key": "value",
				},
				TargetsRegex: map[string]string{
					"key": "value(.)+",
				},
				Sources: map[string]string{
					"key": "value",
				},
				SourcesRegex: map[string]string{
					"key": "value(.)?",
				},
				Labels: []string{
					"label",
					"lebal",
				},
			},
		},
	}
	return alert
}

func ValidConfiguration() *naisiov1.Alert {
	return ConfigurationTemplate(validAction, validDescription, validExpr)
}

func InvalidExpr() *naisiov1.Alert {
	expr := "kube_deployment_status_replicas_unavailable{deployment=\"nais-testapp\"} > <not valid value>"
	return ConfigurationTemplate(validAction, validDescription, expr)
}

func InvalidAction() *naisiov1.Alert {
	action := "kubectl describe pod -l app=nais-testapp -n {{ $asdf.namespace }}"
	return ConfigurationTemplate(action, validDescription, validExpr)
}

func InvalidDescription() *naisiov1.Alert {
	description := "Oh noes, it looks like Nais-testapp is down in {{ namespace }}"
	return ConfigurationTemplate(validAction, description, validExpr)
}
