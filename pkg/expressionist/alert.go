package expressionist

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v2"
)

type alert struct {
	MetaData struct {
		Name string
	} `json:"metadata"`

	Spec alertSpec
}

type alertSpec struct {
	Alerts []rule
}

type rule struct {
	Alert       string
	Expr        string
	description string
}

type rules struct {
	Groups []group
}

type group struct {
	Name  string
	Rules []rule
}

func ParseDescription(applied string) error {
	var alert alert
	err := yaml.Unmarshal([]byte(applied), &alert)
	if err != nil {
		return fmt.Errorf("failed while unmarshling alertmanager.yml: %s", err)
	}

	for _, rule := range alert.Spec.Alerts {
		if strings.HasPrefix(rule.description, "{{") {
			return fmt.Errorf("missing quotation around 'description'")
		}
	}

	return nil
}

func ParseExpr(applied string) (string, error) {
	var alert alert
	err := yaml.Unmarshal([]byte(applied), &alert)
	if err != nil {
		return "", fmt.Errorf("failed while unmarshling alertmanager.yml: %s", err)
	}

	rules := rules{
		Groups: []group{
			{
				Name:  "expressionist",
				Rules: alert.Spec.Alerts,
			},
		},
	}

	filename := strings.Split(alert.MetaData.Name, ".")[0]
	filepath := fmt.Sprintf("/tmp/%s.yaml", filename)
	err = writeRulesToFile(filepath, rules)
	if err != nil {
		return "", fmt.Errorf("%s: %s", filename, err)
	}

	return validateRulesInFile(filepath), nil
}

func writeRulesToFile(filepath string, rules rules) error {
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
