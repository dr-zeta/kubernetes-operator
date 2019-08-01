package notifier

import (
	"fmt"
	"github.com/go-logr/logr"
	"net/http"

	"github.com/jenkinsci/kubernetes-operator/pkg/apis/jenkins/v1alpha2"
	"github.com/jenkinsci/kubernetes-operator/pkg/log"

	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// LogWarn is warning log entry
	LogWarn LoggingLevel = "warn"

	// LogInfo is info log entry
	LogInfo LoggingLevel = "info"

	titleText                  = "Operator reconciled."
	messageFieldName           = "Message"
	loggingLevelFieldName      = "Logging Level"
	crNameFieldName            = "CR Name"
	configurationTypeFieldName = "Configuration Type"
	namespaceFieldName         = "Namespace"
	footerContent              = "Powered by Jenkins Operator <3"
)

var (
	testConfigurationType = "test-configuration"
	testCrName            = "test-cr"
	testNamespace         = "test-namespace"
	testMessage           = "test-message"
	testMessageVerbose    = "detail-test-message"
	testLoggingLevel      = LogWarn

	client = http.Client{}
)

// StatusColor is useful for better UX
type StatusColor string

// LoggingLevel is type for selecting different logging levels
type LoggingLevel string

// Information represents details about operator status
type Information struct {
	ConfigurationType string
	Namespace         string
	CrName            string
	LogLevel          LoggingLevel
	Message           string
	MessageVerbose    string
}

// Notification contains message which will be sent
type Notification struct {
	Jenkins     v1alpha2.Jenkins
	K8sClient   k8sclient.Client
	Logger      logr.Logger
	Information Information
}

// Service is skeleton for additional services
type service interface {
	Send(i *Notification, config v1alpha2.Notification) error
}

// Listen is goroutine that listens for incoming messages and sends it
func Listen(notification chan *Notification) {
	for n := range notification {
		notificationConfig := n.Jenkins.Spec.Notification
		var err error
		var svc service

		if notificationConfig.Slack != (v1alpha2.Slack{}) {
			svc = Slack{}
		} else if notificationConfig.Teams != (v1alpha2.Teams{}) {
			svc = Teams{}
		} else if notificationConfig.Mailgun != (v1alpha2.Mailgun{}) {
			svc = Mailgun{}
		} else {
			n.Logger.V(log.VWarn).Info(fmt.Sprintf("Notification service in `%s` not found or not defined", notificationConfig.Name))
			continue
		}

		err = notify(svc, n, notificationConfig)

		if err != nil {
			n.Logger.V(log.VWarn).Info(fmt.Sprintf("Failed to send notifications. %+v", err))
		} else {
			n.Logger.V(log.VDebug).Info("Sent notification")
		}
	}
}

func getStatusColor(logLevel LoggingLevel, svc service) StatusColor {
	switch svc.(type) {
	case Slack:
		switch logLevel {
		case LogInfo:
			return "#439FE0"
		case LogWarn:
			return "danger"
		default:
			return "#c8c8c8"
		}
	case Teams:
		switch logLevel {
		case LogInfo:
			return "439FE0"
		case LogWarn:
			return "E81123"
		default:
			return "C8C8C8"
		}
	case Mailgun:
		switch logLevel {
		case LogInfo:
			return "blue"
		case LogWarn:
			return "red"
		default:
			return "gray"
		}
	default:
		return "#c8c8c8"
	}
}

func notify(svc service, n *Notification, nc v1alpha2.Notification) error {
	var err error
	switch s := svc.(type) {
	case Slack:
		err = s.Send(n, nc)
	case Teams:
		err = s.Send(n, nc)
	case Mailgun:
		err = s.Send(n, nc)
	}

	return err
}
