package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/jamiealquiza/envy"
	"github.com/pkg/errors"
	whhttp "github.com/slok/kubewebhook/pkg/http"
	"github.com/slok/kubewebhook/pkg/log"
	mutatingwh "github.com/slok/kubewebhook/pkg/webhook/mutating"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type ServiceAnnotator struct {
	rules []Rule
}

func NewServiceAnnotator(rules []Rule) mutatingwh.Mutator {
	return &ServiceAnnotator{
		rules: rules,
	}
}

func (s *ServiceAnnotator) Mutate(ctx context.Context, obj metav1.Object) (bool, error) {
	service, ok := obj.(*corev1.Service)
	if !ok {
		return false, nil
	}

	fmt.Println("asdasds")

	if service.Annotations == nil {
		service.Annotations = make(map[string]string)
	}

	for _, rule := range s.rules {
		if labels.Set(rule.Selector).AsSelector().Matches(labels.Set(service.Labels)) {
			for key, value := range rule.Annotations {
				service.Annotations[key] = value
			}
		}
	}

	return false, nil
}

type Rule struct {
	Selector    map[string]string `json:"selector" yaml:"selector"`
	Annotations map[string]string `json:"annotations" yaml:"annotations"`
}

func parseConfig(path string) ([]Rule, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	rules := make([]Rule, 0)
	err = yaml.Unmarshal(content, &rules)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("unable to open rules file: %q", path))
	}
	logger.Infof("There are %d active rules", len(rules))
	return rules, nil
}

type config struct {
	certFile   string
	keyFile    string
	rulesFile  string
	listenAddr string
	debug      bool
}

func initFlags() *config {
	cfg := &config{}

	flag.StringVar(&cfg.certFile, "tls-cert-file", "", "TLS certificate file")
	flag.StringVar(&cfg.keyFile, "tls-key-file", "", "TLS key file")
	flag.StringVar(&cfg.rulesFile, "rules-file", "rules.yaml", "Rule file path")
	flag.StringVar(&cfg.listenAddr, "listen-addr", ":8080", "Listen address")
	flag.BoolVar(&cfg.debug, "debug", false, "Log debug messages")
	envy.Parse("KSA")
	flag.Parse()
	return cfg
}

var logger log.Logger

func main() {

	cfg := initFlags()
	logger = &log.Std{Debug: cfg.debug}
	logger.Infof("Reading rules from: %q", cfg.rulesFile)
	rules, err := parseConfig(cfg.rulesFile)
	if err != nil {
		logger.Errorf("error parsing configuration: %s", err.Error())
		os.Exit(1)
	}

	ServiceMutator := NewServiceAnnotator(rules)
	ServiceMutatorCfg := mutatingwh.WebhookConfig{
		Name: "serviceAnnotate",
		Obj:  &corev1.Service{},
	}

	wh, err := mutatingwh.NewWebhook(ServiceMutatorCfg, ServiceMutator, nil, nil, logger)
	if err != nil {
		logger.Errorf("error creating webhook: %s", err.Error())
		os.Exit(1)
	}

	whHandler, err := whhttp.HandlerFor(wh)
	if err != nil {
		logger.Errorf("error creating webhook handler: %s", err.Error())
		os.Exit(1)
	}
	if cfg.certFile != "" && cfg.keyFile != "" {
		logger.Infof("Listening TLS on %s", cfg.listenAddr)
		err = http.ListenAndServeTLS(cfg.listenAddr, cfg.certFile, cfg.keyFile, whHandler)
	} else {
		logger.Infof("Listening unsecure on %s", cfg.listenAddr)
		err = http.ListenAndServe(cfg.listenAddr, whHandler)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error serving webhook: %s", err)
		os.Exit(1)
	}
}
