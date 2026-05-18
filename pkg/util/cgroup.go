package util

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

const mountInfoPath string = "/proc/self/mountinfo"
const hostDeviceFlag string = "/etc/hosts"
const cgroupInfoPath string = "/proc/self/cgroup"

type MountInfo struct {
	Device            string
	Fstype            string
	Root              string
	MountPoint        string
	Opts              []string
	Major             string
	Minor             string
	SuperBlockOptions []string
}

func (mi MountInfo) String() string {
	optStr := strings.Join(mi.Opts, ",")
	superBlockOptionsStr := strings.Join(mi.SuperBlockOptions, ",")
	return fmt.Sprintf("%s:%s %s %s %s - %s %s %s", mi.Major, mi.Minor, mi.Root, mi.MountPoint, optStr, mi.Fstype, mi.Device, superBlockOptionsStr)
}

func FindTargetDeviceID(mi *MountInfo) bool {
	if mi.MountPoint == hostDeviceFlag {
		log.Printf("found host blockDeviceId Major: %s Minor: %s\n", mi.Major, mi.Minor)
		return true
	}
	return false
}

func GetMountInfo() ([]MountInfo, error) {
	f, err := os.Open(mountInfoPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var ret []string

	r := bufio.NewReader(f)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			break
		}
		ret = append(ret, strings.Trim(line, "\n"))
	}
	mountInfos := make([]MountInfo, len(ret))

	for _, r := range ret {
		parts := strings.Split(r, " - ")
		if len(parts) != 2 {
			return nil, fmt.Errorf("found invalid mountinfo line in file %s: %s ", mountInfoPath, r)
		}
		mi := MountInfo{}

		fields := strings.Fields(parts[0])
		blockId := strings.Split(fields[2], ":")
		if len(blockId) != 2 {
			return nil, fmt.Errorf("found invalid mountinfo line in file %s: %s ", mountInfoPath, r)
		}
		mi.Major = blockId[0]
		mi.Minor = blockId[1]
		mi.Root = fields[3]
		mi.MountPoint = fields[4]
		mi.Opts = strings.Split(fields[5], ",")

		fields = strings.Fields(parts[1])
		if len(fields) <= 1 || len(fields) > 3 {
			return nil, fmt.Errorf("found invalid mountinfo line in file %s: %s ", mountInfoPath, r)
		}

		mi.Fstype = fields[0]

		if len(fields) == 2 {
			mi.Device = ""
			mi.SuperBlockOptions = strings.Split(fields[1], ",")
		} else {
			mi.Device = fields[1]
			mi.SuperBlockOptions = strings.Split(fields[2], ",")
		}

		mountInfos = append(mountInfos, mi)
	}

	return mountInfos, err
}

func MakeDev(major, minor string) int {
	ret1, err := strconv.ParseInt(major, 10, 64)
	if err != nil {
		log.Printf("convert major number to int64 err: %v\n", err)
		return 0
	}
	ret2, err := strconv.ParseInt(minor, 10, 64)
	if err != nil {
		log.Printf("convert minor number to int64 err: %v\n", err)
		return 0
	}

	return int(((ret1 & 0xfff) << 8) | (ret2 & 0xff) | ((ret1 &^ 0xfff) << 32) | ((ret2 & 0xfffff00) << 12))
}

func SetBlockAccessible(path string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_SYNC, 0200)
	if err != nil {
		return fmt.Errorf("open devices.allow failed. %v\n", err)
	}
	defer f.Close()

	l, err := f.Write([]byte("a"))
	if err != nil {
		return fmt.Errorf("write devices.allow failed. %v\n", err)
	}

	if l != 1 {
		return fmt.Errorf("write \"a\" to devices.allow failed.\n")
	}
	log.Printf("set all block device accessible success.\n")

	return nil
}

func GetKernelVersion() ([]int, error) {
	utsInfo := &unix.Utsname{}
	err := unix.Uname(utsInfo)
	if err != nil {
		return nil, err
	}
	relStr := string(utsInfo.Release[:])
	relIdx := strings.Index(relStr, "-")
	if relIdx == -1 {
		return nil, errors.New("unknown internal error when executing uname")
	}
	ret := make([]int, 3)
	for _, v := range strings.Split(relStr[:relIdx], ".") {
		verData, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		ret = append(ret, verData)
	}
	return ret, nil
}

func GetCgroupVersion() (int, error) {
	_, err := os.Stat("/sys/fs/cgroup/cgroup.controllers")
	if err == nil {
		return 2, nil
	}
	if strings.Contains(err.Error(), "no such file or directory") {
		return 1, nil
	}
	return -1, err
}

type CgroupInfo struct {
	HierarchyID   int
	ControllerLst string
	CgroupPath    string
	OriginalInfo  string
}

func GetAllCGroup() ([]CgroupInfo, error) {
	return GetCgroup(0)
}

func GetCgroup(pid int) ([]CgroupInfo, error) {
	var cginfo []CgroupInfo
	var pidStr string

	if pid == 0 {
		pidStr = "self"
	} else {
		pidStr = fmt.Sprint(pid)
	}

	cgroupInfoPath := fmt.Sprintf("/proc/%s/cgroup", pidStr)
	datafd, err := os.Open(cgroupInfoPath)
	if err != nil {
		return nil, err
	}
	defer datafd.Close()

	sc := bufio.NewScanner(datafd)
	for sc.Scan() {
		originalInfo := sc.Text()
		singleCG := strings.Split(strings.TrimSuffix(originalInfo, "\n"), ":")
		hID, err := strconv.Atoi(singleCG[0])
		if err != nil {
			return nil, err
		}
		cginfo = append(cginfo, CgroupInfo{hID, singleCG[1], singleCG[2], originalInfo})
	}

	return cginfo, nil
}

func GetAllCGroupSubSystem() ([]string, error) {
	cgSyses, err := GetAllCGroup()
	if err != nil {
		return nil, err
	}
	var syses []string
	for _, v := range cgSyses {
		syses = append(syses, v.ControllerLst)
	}
	return syses, nil
}
