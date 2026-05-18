package dockerd_api

import (
	"fmt"
	"k8strike/pkg/util"
	"log"
)

func UcurlToolApi(args []string) {
	if len(args) < 3 || len(args) > 4 {
		log.Fatal("invalid input args, Example: ./k8strike tool ucurl get /var/run/docker.sock http://127.0.0.1/info [<body>]")
	}
	body := ""
	if len(args) == 4 {
		body = args[3]
	}
	ans, err := util.UnixHttpSend(args[0], args[1], args[2], body)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("response:")
	fmt.Println(ans)
}

func DcurlToolApi(args []string) {
	if len(args) < 2 || len(args) > 3 {
		log.Fatal("invalid input args, Example: ./k8strike tool dcurl get http://127.0.0.1:2375/info [<body>]")
	}
	body := ""
	if len(args) == 3 {
		body = args[2]
	}
	ans, err := util.HttpSendJson(args[0], args[1], body)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("response:")
	fmt.Println(ans)
}
