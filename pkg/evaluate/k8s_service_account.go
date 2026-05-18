package evaluate

import (
	"fmt"
	"log"
	"strings"

	"k8strike/pkg/conf"
	"k8strike/pkg/tool/kubectl"
)

func CheckPrivilegedK8sServiceAccount(tokenPath string) bool {
	resp, err := kubectl.ServerAccountRequest(
		kubectl.K8sRequestOption{
			TokenPath: "",
			Server:    "",
			Api:       "/apis",
			Method:    "get",
			PostData:  "",
			Anonymous: false,
		})
	if err != nil {
		fmt.Println(err)
		return false
	}
	if len(resp) > 0 && strings.Contains(resp, "APIGroupList") {
		fmt.Println("\tservice-account is available")

		log.Println("trying to list namespaces")
		resp, err := kubectl.ServerAccountRequest(
			kubectl.K8sRequestOption{
				TokenPath: "",
				Server:    "",
				Api:       "/api/v1/namespaces",
				Method:    "get",
				PostData:  "",
				Anonymous: false,
			})
		if err != nil {
			fmt.Println(err)
			return false
		}
		if len(resp) > 0 && strings.Contains(resp, "kube-system") {
			fmt.Println("\tsuccess, the service-account have a high authority.")
			fmt.Println("\tnow you can make your own request to takeover the entire k8s cluster with `./k8strike tool kcurl` command\n\tgood luck and have fun.")
			return true
		} else {
			fmt.Println("\tfailed")
			fmt.Println("\tresponse:" + resp)
			return false
		}
	} else {
		fmt.Println("\tservice-account is not available")
		fmt.Println("\tresponse:" + resp)
		return false
	}
}

func init() {
	RegisterSimpleCheck(
		CategoryK8sServiceAccount,
		"k8s.privileged_service_account",
		"Check Kubernetes service account privileges",
		func() {
			CheckPrivilegedK8sServiceAccount(conf.K8sSATokenDefaultPath)
		},
	)
}
