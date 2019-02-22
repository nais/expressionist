package expressionist

import (
	"fmt"
	"github.com/golang/glog"
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
	glog.Infof("We got a request: %s", request)
	// default deny
	return Response{Allowed: false, Reason: fmt.Sprint("hello world")}
}
