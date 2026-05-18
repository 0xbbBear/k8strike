package main

import (
	"k8strike/pkg/cli"
	_ "k8strike/pkg/exploit"
)

func main() {
	cli.Execute()
}
