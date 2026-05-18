package util

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"

	"k8strike/pkg/errors"
)

func ByteToString(orig []byte) string {
	n := -1
	l := -1
	for i, b := range orig {
		if l == -1 && b == 0 {
			continue
		}
		if l == -1 {
			l = i
		}

		if b == 0 {
			break
		}
		n = i + 1
	}
	if n == -1 {
		return string(orig)
	}
	return string(orig[l:n])
}

func RandString(n int) string {
	const (
		letterBytes   = "abcde1fghij2klmno3pqrst4uvwxy5zABCD6EFGHI7JKLMN8OPQRS9TUVWX9YZ"
		letterIdxBits = 6
		letterIdxMask = 1<<letterIdxBits - 1
		letterIdxMax  = 63 / letterIdxBits
	)
	sb := strings.Builder{}
	sb.Grow(n)
	for i, cache, remain := n-1, rand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return sb.String()
}

func RemoveDuplicateElement(addrs []string) []string {
	result := make([]string, 0, len(addrs))
	temp := map[string]struct{}{}
	for _, item := range addrs {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

func dataFromSliceOrFile(data []byte, file string) ([]byte, error) {
	if len(data) > 0 {
		return data, nil
	}
	if len(file) > 0 {
		fileData, err := os.ReadFile(file)
		if err != nil {
			return []byte{}, err
		}
		return fileData, nil
	}
	return nil, nil
}

func ShellExec(shellPath string) error {
	var command = shellPath
	if strings.HasPrefix(shellPath, "/") {
		command = shellPath
	} else {
		command = fmt.Sprintf("./%s .", shellPath)
	}
	cmd := exec.Command("/bin/bash", "-c", command)

	output, err := cmd.Output()
	if err != nil {
		return &errors.K8strikeRuntimeError{Err: err, CustomMsg: fmt.Sprintf("Execute Shell:%s failed", command)}
	}
	fmt.Printf("Execute Shell:%s finished with output:\n%s", command, string(output))
	return nil
}

func StringContains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func IntContains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func DistinctStrArr(s []string) []string {
	distinctMap := make(map[string]bool)
	var result []string

	for _, item := range s {
		if _, exists := distinctMap[item]; !exists {
			distinctMap[item] = true
			result = append(result, item)
		}
	}

	return result
}
