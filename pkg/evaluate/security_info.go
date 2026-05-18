package evaluate

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"log"
	"os"
	"strings"
)

var namespaceTypes = []string{"cgroup", "ipc", "mnt", "net", "pid", "uts"}

func CheckNamespaceIsolation() {
	log.Println("Namespace isolation status:")
	for _, ns := range namespaceTypes {
		initTarget, err1 := os.Readlink(fmt.Sprintf("/proc/1/ns/%s", ns))
		selfTarget, err2 := os.Readlink(fmt.Sprintf("/proc/self/ns/%s", ns))
		if err1 != nil || err2 != nil {
			log.Printf("\t%s: unable to read namespace links", ns)
			continue
		}
		if initTarget != selfTarget {
			fmt.Printf("\t%s: isolated (%s)\n", ns, selfTarget)
		} else {
			fmt.Printf("\t%s: NOT isolated (shared with host, %s)\n", ns, selfTarget)
		}
	}
}

func CheckSeccompStatus() {
	data, err := os.ReadFile("/proc/self/status")
	if err != nil {
		log.Printf("seccomp: unable to read /proc/self/status: %v", err)
		return
	}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Seccomp:") {
			parts := strings.Fields(line)
			if len(parts) < 2 {
				log.Println("seccomp: malformed Seccomp line")
				return
			}
			switch parts[1] {
			case "0":
				log.Println("Seccomp: disabled")
			case "1":
				log.Println("Seccomp: strict mode (1)")
			case "2":
				log.Println("Seccomp: filter mode (2)")
			default:
				log.Printf("Seccomp: unknown value %s", parts[1])
			}
			return
		}
	}
	log.Println("Seccomp: field not found in /proc/self/status (kernel may not support Seccomp)")
}

func CheckSeccompKernelSupport() {
	data, err := os.ReadFile("/proc/self/status")
	if err != nil {
		log.Printf("seccomp: unable to read /proc/self/status: %v", err)
		return
	}
	if strings.Contains(string(data), "Seccomp:") {
		log.Println("Seccomp: kernel supports Seccomp")
	} else {
		log.Println("Seccomp: kernel does NOT support Seccomp")
	}

	if val, ok := readKernelConfigOption("CONFIG_SECCOMP"); ok {
		log.Printf("Seccomp: kernel config CONFIG_SECCOMP=%s", val)
	}
}

func CheckSELinux() {
	enforceFile := "/sys/fs/selinux/enforce"
	data, err := os.ReadFile(enforceFile)
	if err != nil {
		log.Println("SELinux: not detected (no selinuxfs)")
		return
	}
	switch strings.TrimSpace(string(data)) {
	case "1":
		log.Println("SELinux: enforcing")
	case "0":
		log.Println("SELinux: permissive (loaded but not enforcing)")
	default:
		log.Printf("SELinux: unexpected enforce value %q", strings.TrimSpace(string(data)))
	}

	if label, err := os.ReadFile("/proc/self/attr/current"); err == nil {
		trimmed := strings.TrimRight(string(label), "\x00\n")
		log.Printf("SELinux: container label: %s", trimmed)
	}
}

func CheckAppArmor() {
	if val, ok := readKernelConfigOption("CONFIG_SECURITY_APPARMOR"); ok {
		log.Printf("AppArmor: kernel config CONFIG_SECURITY_APPARMOR=%s", val)
	} else {
		log.Println("AppArmor: kernel config not available")
	}

	if cmdline, err := os.ReadFile("/proc/cmdline"); err == nil {
		params := string(cmdline)
		if strings.Contains(params, "apparmor=1") || strings.Contains(params, "security=apparmor") {
			log.Printf("AppArmor: enabled via boot parameters (%s)", strings.TrimSpace(params))
		} else if strings.Contains(params, "apparmor=0") {
			log.Println("AppArmor: disabled via boot parameter apparmor=0")
		} else {
			log.Println("AppArmor: no explicit AppArmor boot parameter found")
		}
	}

	if data, err := os.ReadFile("/sys/module/apparmor/parameters/enabled"); err == nil {
		if strings.TrimSpace(string(data)) == "Y" {
			log.Println("AppArmor: module is enabled (runtime)")
		} else {
			log.Println("AppArmor: module is loaded but disabled (runtime)")
		}
	} else {
		log.Println("AppArmor: module not loaded")
	}

	if label, err := os.ReadFile("/proc/self/attr/current"); err == nil {
		trimmed := strings.TrimRight(string(label), "\x00\n")
		if trimmed == "" || trimmed == "unconfined" {
			log.Println("AppArmor: container is unconfined (no profile attached)")
		} else {
			log.Printf("AppArmor: container profile: %s", trimmed)
		}
	} else {
		log.Println("AppArmor: unable to read container profile")
	}
}

func readKernelConfigOption(key string) (string, bool) {
	if f, err := os.Open("/proc/config.gz"); err == nil {
		defer f.Close()
		if gz, err := gzip.NewReader(f); err == nil {
			defer gz.Close()
			scanner := bufio.NewScanner(gz)
			for scanner.Scan() {
				if val, ok := matchConfigLine(scanner.Text(), key); ok {
					return val, true
				}
			}
			return "", false
		}
	}

	uname, err := os.ReadFile("/proc/sys/kernel/osrelease")
	if err != nil {
		return "", false
	}
	configPath := "/boot/config-" + strings.TrimSpace(string(uname))
	f2, err := os.Open(configPath)
	if err != nil {
		return "", false
	}
	defer f2.Close()
	scanner := bufio.NewScanner(f2)
	for scanner.Scan() {
		if val, ok := matchConfigLine(scanner.Text(), key); ok {
			return val, true
		}
	}
	return "", false
}

func matchConfigLine(line, key string) (string, bool) {
	if strings.HasPrefix(line, key+"=") {
		return strings.TrimPrefix(line, key+"="), true
	}
	if line == "# "+key+" is not set" {
		return "n", true
	}
	return "", false
}

func init() {
	RegisterSimpleCheck(CategorySecurity, "security.namespace_isolation", "Check container namespace isolation", CheckNamespaceIsolation)
	RegisterSimpleCheck(CategorySecurity, "security.seccomp_status", "Check Seccomp status", CheckSeccompStatus)
	RegisterSimpleCheck(CategorySecurity, "security.seccomp_support", "Check kernel Seccomp support", CheckSeccompKernelSupport)
	RegisterSimpleCheck(CategorySecurity, "security.selinux", "Check SELinux status", CheckSELinux)
	RegisterSimpleCheck(CategorySecurity, "security.apparmor", "Check AppArmor status and container profile", CheckAppArmor)
}
