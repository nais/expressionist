package expressionist

import (
	"fmt"
	"github.com/prometheus/common/log"
	"io/ioutil"
	"os/exec"

	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type alert struct {
	MetaData struct {
		Name string
	} `json:metadata`

	Spec alertSpec
}

type alertSpec struct {
	Alerts []rule
}

type rule struct {
	Alert string
	Expr  string
}

type rules struct {
	Groups []group
}

type group struct {
	Name  string
	Rules []rule
}

func ParseAlert(alertObject metav1.Object) (string, error) {
	last := alertObject.GetAnnotations()["kubectl.kubernetes.io/last-applied-configuration"]
	var alert alert
	err := yaml.Unmarshal([]byte(last), &alert)
	if err != nil {
		return "", fmt.Errorf("failed while unmarshling alertmanager.yml: %s", err)
	}
	log.Infof("Parsing alerts for %s", alert.MetaData.Name)

	rules := rules{
		Groups: []group{
			{
				Name:  "expressionist",
				Rules: alert.Spec.Alerts,
			},
		},
	}

	err = writeRulesToFile(rules)
	if err != nil {
		return "", err
	}

	return validateRulesInFile(), nil
}

func writeRulesToFile(rules rules) error {
	data, err := yaml.Marshal(&rules)
	if err != nil {
		return fmt.Errorf("failed while marshaling rules to file: %s", err)
	}

	err = ioutil.WriteFile("/tmp/rules.yaml", data, 0644)
	if err != nil {
		return fmt.Errorf("failed while writing rules to file: %s", err)
	}
	return nil
}

func validateRulesInFile() string {
	tool := "promtool"
	args := []string{"check", "rules", "/tmp/rules.yaml"}
	cmd := exec.Command(tool, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Infof("Promtool (%s):\n%s", err, output)
		return string(output)
	}

	return ""
}
