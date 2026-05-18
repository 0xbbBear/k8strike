package evaluate

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"k8strike/pkg/errors"
	"k8strike/pkg/util"
)

func checkClose(c io.Closer) {
	if err := c.Close(); err != nil {
		panic(&errors.K8strikeRuntimeError{Err: err})
	}
}

func MountEscape() {

	mounts, _ := util.GetMountInfo()

	for _, m := range mounts {

		if m.Major == "" {
			continue
		}

		if strings.Contains(m.Device, "/") || strings.Contains(m.Fstype, "ext") {
			matched, _ := regexp.MatchString("/kubelet/|/dev/[\\w-]*?\\blog$|/etc/host[\\w]*?$|/etc/[\\w]*?\\.conf$", m.Root)
			if !matched {
				m.Root = util.RedBold.Sprint(m.Root)
				m.Fstype = util.RedBold.Sprint(m.Fstype)
			}
		}

		if m.Device == "lxcfs" && util.StringContains(m.Opts, "rw") {
			fmt.Printf("Find mounted lxcfs with rw flags, run `%s` or `%s` to escape container!\n", util.RedBold.Sprint("k8strike run lxcfs-rw"), util.RedBold.Sprint("k8strike run lxcfs-rw-cgroup"))
			m.Device = util.RedBold.Sprint(m.Device)
			m.MountPoint = util.RedBold.Sprint(m.Device)
		}

		fmt.Println(m.String())

	}
}

func init() {
	RegisterSimpleCheck(CategoryMounts, "mounts.escape", "Inspect mount escape opportunities", MountEscape)
}
