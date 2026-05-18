package evaluate

import (
	"log"
	"os"
	"os/user"

	"k8strike/pkg/conf"
	"k8strike/pkg/util"
	"github.com/shirou/gopsutil/v3/host"
)

func BasicSysInfo() {
	dir, err := os.Getwd()
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("current dir:", dir)

	u, err := user.Current()
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("current user:", u.Username, "uid:", u.Uid, "gid:", u.Gid, "home:", u.HomeDir)

	hostname, err := os.Hostname()
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("hostname:", hostname)

	kversion, _ := host.KernelVersion()
	platform, family, osversion, _ := host.PlatformInformation()
	log.Println(family, platform, osversion, "kernel:", kversion)

}

func FindSidFiles() {

	var setuidfiles []string

	for _, dir := range conf.DefaultPathEnv {
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, file := range files {
			if file.Type()&os.ModeSetuid != 0 {
				setuidfiles = append(setuidfiles, dir+"/"+file.Name())
			}

		}
	}

	if len(setuidfiles) > 0 {
		util.PrintItemKey("Setuid files found:", false)
		for _, file := range setuidfiles {
			util.PrintItemValue(file, true)
		}
	}
}

func CommandAllow() {
}

func ASLR() {
	var ASLRSetting = "/proc/sys/kernel/randomize_va_space"

	data, err := os.ReadFile(ASLRSetting)
	if err != nil {
		log.Printf("err found while open %s: %v\n", RouteLocalNetProcPath, err)
		return
	}
	log.Printf("/proc/sys/kernel/randomize_va_space file content: %s", string(data))

	if string(data) == "0" {
		log.Println("ASLR is disabled.")
	} else {
		log.Println("ASLR is enabled.")
	}

}

func init() {
	RegisterSimpleCheck(CategorySystemInfo, "system.basic_info", "Collect basic system information", BasicSysInfo)
	RegisterSimpleCheck(CategorySystemInfo, "system.setuid_files", "Search for setuid binaries", FindSidFiles)
	RegisterSimpleCheck(CategoryASLR, "system.aslr", "Inspect ASLR configuration", ASLR)
}
