package evaluate

import (
	"os/exec"
	"strings"

	"k8strike/pkg/conf"
	"k8strike/pkg/util"
)

func kernelExploitSuggester() {
	script := conf.KernelExploitScript
	_, err := exec.LookPath("bash")
	if err != nil {
		return
	}

	util.PrintItemValueWithKeyOneLine("refer", "https://github.com/mzet-/linux-exploit-suggester", false)
	output, err := exec.Command("bash", "-c", script).Output()
	if err != nil {
		return
	}

	indexs := make([]int, 0)
	lines := strings.Split(string(output), "\n")
	for index, line := range lines {
		if strings.Contains(line, "[CVE") {
			indexs = append(indexs, index)
		}
	}

	for _, index := range indexs {
		for i := index; i < index+10; i++ {
			if i >= len(lines) {
				break
			}

			if i != index && strings.Contains(lines[i], "[CVE") {
				break
			}

			util.PrintOrignal(lines[i])
		}
	}

}

func init() {
	RegisterSimpleCheck(CategoryKernel, "kernel.exploits", "Suggest applicable kernel exploits", kernelExploitSuggester)
}
