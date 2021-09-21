package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	naisiov1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	log "github.com/sirupsen/logrus"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	"github.com/nais/expressionist/pkg/expressionist"
	"github.com/nais/expressionist/pkg/metrics"
	"github.com/nais/expressionist/pkg/version"
)

// Config contains the server (the webhook) cert and key.
type Config struct {
	LogFormat string
	LogLevel  string
	CertFile  string
	KeyFile   string
}

func DefaultConfig() *Config {
	return &Config{
		LogFormat: "text",
		LogLevel:  "info",
		CertFile:  "/cert/cert.pem",
		KeyFile:   "/cert/key.pem",
	}
}

var config = DefaultConfig()

func (c *Config) addFlags() {
	flag.StringVar(&c.LogFormat, "log-format", c.LogFormat, "Log format, either 'json' or 'text'")
	flag.StringVar(&c.LogLevel, "log-level", c.LogLevel, "Logging verbosity level")
	flag.StringVar(&c.CertFile, "cert", c.CertFile, "File containing the x509 certificate for HTTPS")
	flag.StringVar(&c.KeyFile, "key", c.KeyFile, "File containing the x509 private key")
}

func genericErrorResponse(format string, a ...interface{}) *v1beta1.AdmissionResponse {
	return &v1beta1.AdmissionResponse{
		Allowed: false,
		Result: &metav1.Status{
			Message: fmt.Sprintf(format, a...),
		},
	}
}

func decode(raw []byte) (*naisiov1.Alert, error) {
	alert := &naisiov1.Alert{}

	deserializer := serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer()
	if _, _, err := deserializer.Decode(raw, nil, alert); err != nil {
		return nil, fmt.Errorf("while decoding Kubernetes resource: %s", err)
	}

	return alert, nil
}

func admitCallback(admissionReview v1beta1.AdmissionReview) (*v1beta1.AdmissionResponse, error) {
	if admissionReview.Request == nil {
		return nil, fmt.Errorf("admission review request is empty")
	}

	response := &v1beta1.AdmissionResponse{
		Allowed: true,
		Result: &metav1.Status{},
	}

	kind := admissionReview.Request.Kind.Kind
	if kind != "Alert" {
		return response, nil
	}

	alert, err := decode(admissionReview.Request.Object.Raw)
	if err != nil {
		return nil, fmt.Errorf("while decoding resource: %s", err)
	}
	req := expressionist.Request{
		SubmittedResource: alert,
	}

	result := expressionist.Allowed(req)
	response.Allowed = result.Allowed
	response.Result = &metav1.Status{
		Message: result.Reason,
	}

	return response, nil
}

func reply(request *http.Request) (*v1beta1.AdmissionReview, error) {
	contentType := request.Header.Get("Content-Type")
	if contentType != "application/json" {
		return nil, fmt.Errorf("contentType=%s, expect application/json", contentType)
	}

	var reviewResponse *v1beta1.AdmissionResponse
	admissionReview := v1beta1.AdmissionReview{}

	data, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return nil, fmt.Errorf("while reading admission request: %s", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	err = decoder.Decode(&admissionReview)
	if err == nil {
		reviewResponse, err = admitCallback(admissionReview)
	}

	if err != nil {
		reviewResponse = genericErrorResponse(err.Error())
	}

	reviewResponse.UID = admissionReview.Request.UID

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
		metrics.Validations.Inc()
	} else {
		metrics.Failed.Inc()
	}

	encoder := json.NewEncoder(w)
	err = encoder.Encode(review)
	if err != nil {
		log.Errorf("while sending review response: %s", err)
	}
}

func configTLS(config Config) (*tls.Config, error) {
	sCert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("while loading certificate and key file: %s", err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{sCert},
	}, nil
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
	config.addFlags()
	flag.Parse()

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

	tlsConfig, err := configTLS(*config)
	if err != nil {
		return fmt.Errorf("while setting up TLS: %s", err)
	}

	go metrics.Serve(":8080", "/metrics", "/isReady", "/isAlive")

	http.HandleFunc("/", serve)
	server := &http.Server{
		Addr:      ":8443",
		TLSConfig: tlsConfig,
	}
	err = server.ListenAndServeTLS("", "")
	if err != nil {
		return fmt.Errorf("while starting server: %s", err)
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
