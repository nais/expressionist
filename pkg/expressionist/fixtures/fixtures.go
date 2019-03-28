package fixtures

var ValidConfiguration = `{
  "apiVersion": "nais.io/v1alpha1",
  "kind": "Alert",
  "metadata": {
    "annotations": {},
    "labels": {
      "team": "aura"
    },
    "name": "nais-testapp"
  },
  "spec": {
    "alerts": [
      {
        "action": "kubectl describe pod -l app=nais-testapp",
        "alert": "Nais-testapp unavailable",
        "description": "Oh noes, it looks like Nais-testapp is down",
        "documentation": "https://github.com/navikt/aura-doc/naisvakt/alerts.md#app_unavailable",
        "expr": "kube_deployment_status_replicas_unavailable{deployment=\"nais-testapp\"} > 0",
        "for": "2m",
        "severity": "critical",
        "sla": "respond within 1h, during office hours"
      },
      {
        "action": "kubectl describe pod -l app=nais-testapp",
        "alert": "CoreDNS unavailable",
        "description": "CoreDNS unavailable, there are zero replicas",
        "documentation": "https://github.com/navikt/aura-doc/naisvakt/alerts.md#coredns",
        "expr": "kube_deployment_status_replicas_available{namespace=\"kube-system\", deployment=\"coredns\"} == 0",
        "for": "1m",
        "severity": "critical",
        "sla": "respond within 1h, solve within 4h, around the clock"
      }
    ],
    "receivers": {
      "email": {
        "send_resolved": false,
        "to": "auravakt@nav.no"
      },
      "slack": {
        "channel": "#nais-alerts-dev",
        "prependText": "<!here> | "
      }
    }
  }
}`

var NotValidConfiguration = `{
  "apiVersion": "nais.io/v1alpha1",
  "kind": "Alert",
  "metadata": {
    "annotations": {},
    "labels": {
      "team": "aura"
    },
    "name": "nais-testapp"
  },
  "spec": {
    "alerts": [
      {
        "action": "kubectl describe pod -l app=nais-testapp",
        "alert": "Nais-testapp unavailable",
        "description": "Oh noes, it looks like Nais-testapp is down",
        "documentation": "https://github.com/navikt/aura-doc/naisvakt/alerts.md#app_unavailable",
        "expr": "kube_deployment_status_replicas_unavailable{deployment=\"nais-testapp\"} > <notValidValue>",
        "for": "2m",
        "severity": "critical",
        "sla": "respond within 1h, during office hours"
      },
      {
        "action": "kubectl describe pod -l app=nais-testapp",
        "alert": "CoreDNS unavailable",
        "description": "CoreDNS unavailable, there are zero replicas",
        "documentation": "https://github.com/navikt/aura-doc/naisvakt/alerts.md#coredns",
        "expr": "kube_deployment_status_replicas_available{namespace=\"kube-system\", deployment=\"coredns\"} == 0",
        "for": "1m",
        "severity": "critical",
        "sla": "respond within 1h, solve within 4h, around the clock"
      }
    ],
    "receivers": {
      "email": {
        "send_resolved": false,
        "to": "auravakt@nav.no"
      },
      "slack": {
        "channel": "#nais-alerts-dev",
        "prependText": "<!here> | "
      }
    }
  }
}`
