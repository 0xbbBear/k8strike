package evaluate

import (
	"fmt"
	"log"
	"strings"

	"k8strike/pkg/tool/kubectl"
)

func CheckK8sAnonymousLogin() bool {

	log.Println("checking if api-server allows system:anonymous request.")

	resp, err := kubectl.ServerAccountRequest(
		kubectl.K8sRequestOption{
			TokenPath: "",
			Server:    "",
			Api:       "/",
			Method:    "get",
			PostData:  "",
			Anonymous: true,
		})
	if err != nil {
		fmt.Println(err)
	}

	if strings.Contains(resp, "/api") {
		fmt.Println("\tcongrats, api-server allows anonymous request.")
		log.Println("trying to list namespaces")

		resp, err := kubectl.ServerAccountRequest(
			kubectl.K8sRequestOption{
				TokenPath: "",
				Server:    "",
				Api:       "/api/v1/namespaces",
				Method:    "get",
				PostData:  "",
				Anonymous: true,
			})
		if err != nil {
			fmt.Println(err)
		}
		if len(resp) > 0 && strings.Contains(resp, "kube-system") {
			fmt.Println("\tsuccess, the system:anonymous role have a high authority.")
			fmt.Println("\tnow you can make your own request to takeover the entire k8s cluster with `./k8strike tool kcurl` command\n\tgood luck and have fun.")
			return true
		} else {
			fmt.Println("\tfailed.")
			fmt.Println("\tresponse:" + resp)
			return true
		}
	} else {
		fmt.Println("\tapi-server forbids anonymous request.")
		fmt.Println("\tresponse:" + resp)
		return false
	}
}

func init() {
	RegisterSimpleCheck(
		CategoryK8sAPIServer,
		"k8s.anonymous_login",
		"Attempt anonymous Kubernetes API login",
		func() {
			CheckK8sAnonymousLogin()
		},
	)
}
