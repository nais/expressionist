package expressionist

import (
	"fmt"
	"github.com/prometheus/common/log"
	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KubernetesResource represents any Kubernetes resource with standard object metadata structures.
type KubernetesResource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
}

type Request struct {
	UserInfo             authenticationv1.UserInfo
	ExistingResource     metav1.Object
	SubmittedResource    metav1.Object
	ClusterAdmins        []string
	ServiceUserTemplates []string
}

type Response struct {
	Allowed bool
	Reason  string
}

func Allowed(request Request) Response {
	log.Debugf("We got a request: %s", request)
	output, err := ParseAlert(request.SubmittedResource.GetAnnotations()["kubectl.kubernetes.io/last-applied-configuration"])
	if err != nil {
		log.Error(err)
		return Response{Allowed: false, Reason: fmt.Sprintf("Something went wrong: %s", err)}
	}

	if output != "" {
		return Response{false, fmt.Sprintf("Unvalid expr in alert:\n%s", output)}
	}

	return Response{Allowed: false, Reason: fmt.Sprint("Thank you for using Alerterator, we appreciate your business")}
}
