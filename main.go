package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/nais/expressionist/pkg/expressionist"
	"github.com/nais/expressionist/pkg/metrics"
	"github.com/nais/expressionist/pkg/version"
	log "github.com/sirupsen/logrus"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Config contains the server (the webhook) cert and key.
type Config struct {
	LogFormat string
	LogLevel  string
}

func DefaultConfig() *Config {
	return &Config{
		LogFormat: "text",
		LogLevel:  "info",
	}
}

var config = DefaultConfig()

func genericErrorResponse(format string, a ...interface{}) *v1beta1.AdmissionResponse {
	return &v1beta1.AdmissionResponse{
		Allowed: false,
		Result: &metav1.Status{
			Message: fmt.Sprintf(format, a...),
		},
	}
}

func decode(raw []byte) (*expressionist.KubernetesResource, error) {
	k := &expressionist.KubernetesResource{}
	if len(raw) == 0 {
		return nil, nil
	}

	r := bytes.NewReader(raw)
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(k); err != nil {
		return nil, fmt.Errorf("while decoding Kubernetes resource: %s", err)
	}

	return k, nil
}

func admitCallback(ar v1beta1.AdmissionReview) (*v1beta1.AdmissionResponse, error) {
	if ar.Request == nil {
		return nil, fmt.Errorf("admission review request is empty")
	}

	previous, err := decode(ar.Request.OldObject.Raw)
	if err != nil {
		return nil, fmt.Errorf("while decoding old resource: %s", err)
	}

	resource, err := decode(ar.Request.Object.Raw)
	if err != nil {
		return nil, fmt.Errorf("while decoding resource: %s", err)
	}

	req := expressionist.Request{
		UserInfo:          ar.Request.UserInfo,
		ExistingResource:  previous,
		SubmittedResource: resource,
	}

	var selfLink string
	if previous != nil {
		selfLink = previous.GetSelfLink()
	} else if resource != nil {
		selfLink = resource.GetSelfLink()
	}

	if len(selfLink) > 0 {
		log.Infof("Request '%s' from user '%s' in groups %+v", selfLink, ar.Request.UserInfo.Username, ar.Request.UserInfo.Groups)
	} else {
		log.Infof("Request from user '%s' in groups %+v", ar.Request.UserInfo.Username, ar.Request.UserInfo.Groups)
	}

	// These checks are needed in order to avoid a null pointer exception in expressionist.Allowed().
	// Interfaces can be nil checked, but the instances they're pointing to can be nil and
	// still pass through that check.
	if previous == nil {
		req.ExistingResource = nil
	}
	if resource == nil {
		req.SubmittedResource = nil
	}

	log.Tracef("parsed/old: %+v", previous)
	log.Tracef("parsed/new: %+v", resource)

	response := expressionist.Allowed(req)

	reviewResponse := &v1beta1.AdmissionResponse{
		Allowed: response.Allowed,
		Result: &metav1.Status{
			Message: response.Reason,
		},
	}

	fields := log.Fields{
		"user":        ar.Request.UserInfo.Username,
		"groups":      ar.Request.UserInfo.Groups,
		"namespace":   ar.Request.Namespace,
		"operation":   ar.Request.Operation,
		"subresource": ar.Request.SubResource,
		"resource":    selfLink,
	}
	logEntry := log.WithFields(fields)

	if response.Allowed {
		logEntry.Infof("Request allowed: %s", response.Reason)
	} else {
		logEntry.Warningf("Request denied: %s", response.Reason)
	}

	return reviewResponse, nil
}

func reply(r *http.Request) (*v1beta1.AdmissionReview, error) {
	var err error

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		return nil, fmt.Errorf("contentType=%s, expect application/json", contentType)
	}

	var reviewResponse *v1beta1.AdmissionResponse
	ar := v1beta1.AdmissionReview{}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("while reading admission request: %s", err)
	}

	log.Tracef("request: %s", string(data))

	decoder := json.NewDecoder(bytes.NewReader(data))
	err = decoder.Decode(&ar)
	if err == nil {
		reviewResponse, err = admitCallback(ar)
	}

	if err != nil {
		reviewResponse = genericErrorResponse(err.Error())
	}

	reviewResponse.UID = ar.Request.UID

	return &v1beta1.AdmissionReview{
		Response: reviewResponse,
	}, nil
}

func serve(w http.ResponseWriter, r *http.Request) {
	review, err := reply(r)

	if err != nil {
		log.Errorf("while generating review response: %s", err)
	}

	// if there is no review response at this point, we simply cannot provide the API server with a meaningful reply
	// because we couldn't decode a request UID.
	if review == nil {
		return
	}

	if review.Response.Allowed {
		metrics.Admitted.Inc()
	} else {
		metrics.Denied.Inc()
	}

	encoder := json.NewEncoder(w)
	err = encoder.Encode(review)
	if err != nil {
		log.Errorf("while sending review response: %s", err)
	}
}

func textFormatter() log.Formatter {
	return &log.TextFormatter{
		DisableTimestamp: false,
		FullTimestamp:    true,
	}
}

func jsonFormatter() log.Formatter {
	return &log.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
	}
}

func run() error {
	switch config.LogFormat {
	case "json":
		log.SetFormatter(jsonFormatter())
	case "text":
		log.SetFormatter(textFormatter())
	default:
		return fmt.Errorf("log format '%s' is not recognized", config.LogFormat)
	}

	logLevel, err := log.ParseLevel(config.LogLevel)
	if err != nil {
		return fmt.Errorf("while setting log level: %s", err)
	}
	log.SetLevel(logLevel)

	log.Infof("Expressionist v%s (%s)", version.Version, version.Revision)

	go metrics.Serve(":8080", "/metrics", "/ready", "/alive")

	http.HandleFunc("/", serve)
	server := &http.Server{
		Addr: ":8443",
	}
	err = server.ListenAndServeTLS("", "")
	if err != nil {

	}

	log.Info("Shutting down cleanly.")

	return nil
}

func main() {
	err := run()
	if err != nil {
		log.Errorf("Fatal error: %s", err)
		os.Exit(1)
	}
}
