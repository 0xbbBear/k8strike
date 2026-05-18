package util

import (
	"fmt"
	"log"
)

const Colorful = true

func PrintH2(title string) {
	fmt.Printf(BlueBold.Sprint("\n[  ") + GreenBold.Sprint(title) + BlueBold.Sprint("  ]\n"))
}

func PrintItemKey(key string, color bool) {
	key = key + "\n"
	if color {
		log.Printf(YellowBold.Sprint(key))
	} else {
		log.Printf(key)
	}
}

func PrintItemValue(value string, color bool) {
	value = "\t" + value + "\n"
	if color {
		fmt.Printf(RedBold.Sprint(value))
	} else {
		fmt.Printf(value)
	}
}

func PrintItemValueWithKeyOneLine(key, value string, color bool) {
	if color {
		log.Printf("%s: %s", key, GreenBold.Sprint(value))
	} else {
		log.Printf("%s: %s", key, value)
	}
}

func PrintOrignal(out string) {
	fmt.Printf("%s\n", out)
}
