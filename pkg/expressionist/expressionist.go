package expressionist

import (
	"fmt"

	naisiov1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	log "github.com/sirupsen/logrus"
)

type Request struct {
	SubmittedResource *naisiov1.Alert
}

type Response struct {
	Allowed bool
	Reason  string
}

func validateRules(applied *naisiov1.Alert) Response {
	output, err := ValidateRules(applied)
	if err != nil {
		log.Error(err)
		return Response{Allowed: false, Reason: fmt.Sprintf("Something went wrong: %s", err)}
	}

	if output != "" {
		return Response{false, fmt.Sprintf("Invalid rules in alert:\n%s", output)}
	}

	return Response{Allowed: true}
}

func Allowed(request Request) Response {
	log.Debugf("We got a request: %+v", request)
	applied := request.SubmittedResource

	response := validateRules(applied)
	if !response.Allowed {
		return response
	}

	return Response{Allowed: true, Reason: fmt.Sprint("Thank you for using Alerterator, we appreciate your business")}
}
