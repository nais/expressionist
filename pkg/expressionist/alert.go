package expressionist

import (
	"fmt"
	"io/ioutil"
	"os/exec"

	"github.com/nais/alerterator/controllers/rules"
	"github.com/nais/alerterator/utils"
	naisiov1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func ValidateRules(applied *naisiov1.Alert) (string, error) {
	alertRules := rules.CreateAlertRules(applied)
	alertGroups := rules.Groups{
		Groups: []rules.Group{
			{
				Name:  utils.GetCombinedName(applied),
				Rules: alertRules,
			},
		},
	}

	filename := utils.GetCombinedName(applied)
	filepath := fmt.Sprintf("/tmp/%s.yaml", filename)
	err := writeRulesToFile(filepath, alertGroups)
	if err != nil {
		return "", fmt.Errorf("%s: %s", filename, err)
	}

	return validateRulesInFile(filepath), nil
}

func writeRulesToFile(filepath string, rules rules.Groups) error {
	data, err := yaml.Marshal(&rules)
	if err != nil {
		return fmt.Errorf("failed while marshaling rules to file: %s", err)
	}

	err = ioutil.WriteFile(filepath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed while writing rules to file: %s", err)
	}
	return nil
}

func validateRulesInFile(filepath string) string {
	tool := "promtool"
	args := []string{"check", "rules", filepath}
	cmd := exec.Command(tool, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Infof("Promtool (%s):\n%s", err, output)
		return string(output)
	}

	return ""
}
