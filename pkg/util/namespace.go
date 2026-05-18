package util

import (
	"os"
	"strings"

	"k8strike/pkg/errors"
)

func CheckUnpriUserNS() error {

	data, err := os.ReadFile("/proc/sys/kernel/unprivileged_userns_clone")
	if err != nil {
		return &errors.K8strikeRuntimeError{Err: err, CustomMsg: "check prerequisites error."}
	}

	if strings.TrimSuffix(string(data), "\n") != "1" {
		return &errors.K8strikeRuntimeError{Err: nil, CustomMsg: "host os does NOT enable unprivileged user namespace."}
	}

	return nil
}
