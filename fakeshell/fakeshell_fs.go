package fakeshell

import (
	_ "embed"

	"github.com/blang/vfs"
	"github.com/blang/vfs/memfs"
)

//go:embed busybox
var busybox []byte

type FakeDir int

var lfsDirs = []string{
	"/bin",
	"/dev",
	"/etc",
	"/home",
	"/lib",
	"/mnt",
	"/proc",
	"/root",
	"/sbin",
	"/sys",
	"/tmp",
	"/usr/bin",
	"/usr/lib",
	"/var",

	"/dev/shm",
}

var textFiles = map[string]string{
	"/proc/mounts": `/dev/root /rom squashfs ro,relatime 0 0
proc /proc proc rw,nosuid,nodev,noexec,noatime 0 0
sysfs /sys sysfs rw,nosuid,nodev,noexec,noatime 0 0
cgroup2 /sys/fs/cgroup cgroup2 rw,nosuid,nodev,noexec,relatime,nsdelegate 0 0
tmpfs /tmp tmpfs rw,nosuid,nodev,noatime 0 0
/dev/mtdblock6 /overlay jffs2 rw,noatime 0 0
overlayfs:/overlay / overlay rw,noatime,lowerdir=/,upperdir=/overlay/upper,workdir=/overlay/work 0 0
tmpfs /dev tmpfs rw,nosuid,relatime,size=512k,mode=755 0 0
devpts /dev/pts devpts rw,nosuid,noexec,relatime,mode=600,ptmxmode=000 0 0
debugfs /sys/kernel/debug debugfs rw,noatime 0 0
none /sys/fs/bpf bpf rw,nosuid,nodev,noexec,noatime,mode=700 0 0,
`,
	"/proc/partitions": `major minor  #blocks  name

  31        0        192 mtdblock0
  31        1         64 mtdblock1
  31        2         64 mtdblock2
  31        3      16064 mtdblock3
  31        4       2690 mtdblock4
  31        5      13373 mtdblock5
  31        6       9408 mtdblock6
`,
	"/proc/cpuinfo": `system type		: RTL8196E
machine			: Unknown
processor		: 0
cpu model		: 52481
BogoMIPS		: 398.13
tlb_entries		: 32
mips16 implemented	: yes
`}

func MakeFS() vfs.Filesystem {
	fs := memfs.Create()
	for _, dir := range lfsDirs {
		vfs.MkdirAll(fs, dir, 0755)
	}
	vfs.WriteFile(fs, "/bin/busybox", busybox, 0755)
	for path := range fakeshellBinaries {
		vfs.WriteFile(fs, path, busybox, 0755)
	}
	for _, symlink := range busyboxSymlinks {
		fs.Symlink("/bin/busybox", symlink)
	}
	for file, content := range textFiles {
		vfs.WriteFile(fs, file, []byte(content), 0644)
	}
	return fs
}
