package fakeshell

var unameData = map[string]string{
	"m": "mips",
	"n": "papaj-2137XG",
	"r": "5.10.88",
	"s": "Linux",
	"v": "#0 SMP Tue Dec 28 17:04:13 2021",
	"o": "GNU/Linux",
}

var busyboxSymlinks = []string{
	"/bin/adduser", "/bin/ar", "/bin/ash", "/bin/awk", "/bin/brctl", "/bin/bunzip2",
	"/bin/bzcat", "/bin/bzip2", "/bin/cat", "/bin/chmod", "/bin/chpasswd",
	"/bin/cp", "/bin/cpio", "/bin/cut", "/bin/date", "/bin/dd", "/bin/df", "/bin/dirname",
	"/bin/dmesg", "/bin/dpkg", "/bin/dpkg-deb", "/bin/dropbear", "/bin/echo", "/bin/expr",
	"/bin/false", "/bin/free", "/bin/grep", "/bin/gunzip", "/bin/gzip",
	"/bin/halt", "/bin/head", "/bin/hostname", "/bin/ifconfig", "/bin/init", "/bin/ip",
	"/bin/kill", "/bin/killall", "/bin/klogd", "/bin/ln", "/bin/login", "/bin/ls", "/bin/lzmacat",
	"/bin/md5sum", "/bin/mkdir", "/bin/mknod", "/bin/mount", "/bin/mv",
	"/bin/ntpclient", "/bin/passwd", "/bin/pgrep", "/bin/ping", "/bin/poweroff",
	"/bin/ps", "/bin/pwd", "/bin/reboot", "/bin/renice", "/bin/rm", "/bin/route",
	"/bin/rpm", "/bin/rpm2cpio", "/bin/sed", "/bin/sh", "/bin/sleep", "/bin/stty",
	"/bin/syslogd", "/bin/tail", "/bin/tar", "/bin/tftp", "/bin/top", "/bin/true",
	"/bin/udhcpc", "/bin/udhcpc.script", "/bin/umount", "/bin/uncompress",
	"/bin/unlzma", "/bin/unzip", "/bin/uptime", "/bin/vconfig", "/bin/vi",
	"/bin/wc", "/bin/xargs", "/bin/zcat", "/bin/id", "/bin/uname"}

var busyboxHeader = "BusyBox v1.13.4 (2020-04-28 13:57:36 CST) multi-call binary"

var busyboxHelp = busyboxHeader + `
Copyright (C) 1998-2008 Erik Andersen, Rob Landley, Denys Vlasenko
and others. Licensed under GPLv2.
See source distribution for full notice.

Usage: busybox [function] [arguments]...
   or: function [arguments]...

	BusyBox is a multi-call binary that combines many common Unix
	utilities into a single executable.  Most people will create a
	link to busybox for each function they wish to use and BusyBox
	will act like whatever it was invoked as!

Currently defined functions:
	adduser, ar, ash, awk, bunzip2, bzcat, bzip2, cat, chmod, chpasswd,
	cp, cpio, cut, date, dd, df, dirname, dmesg, dpkg, dpkg-deb, echo,
	expr, false, free, grep, gunzip, gzip, halt, head, hostname, id, ifconfig,
	init, ip, kill, killall, klogd, ln, login, ls, lzmacat, md5sum,
	mkdir, mknod, mount, mv, passwd, pgrep, ping, poweroff, ps, pwd,
	reboot, renice, rm, route, rpm, rpm2cpio, sed, sh, sleep, stty,
	syslogd, tail, tar, tftp, top, true, udhcpc, umount, uncompress,
	unlzma, unzip, uptime, vconfig, vi, wc, xargs, zcat
`

var busyboxHelps = map[string]string{
	"cp": busyboxHeader + `

Usage: cp [-arPLHpfinlsTu] SOURCE DEST
or: cp [-arPLHpfinlsu] SOURCE... { -t DIRECTORY | DIRECTORY }

Copy SOURCEs to DEST

	-a	Same as -dpR
	-R,-r	Recurse
	-d,-P	Preserve symlinks (default if -R)
	-L	Follow all symlinks
	-H	Follow symlinks on command line
	-p	Preserve file attributes if possible
	-f	Overwrite
	-i	Prompt before overwrite
	-n	Don't overwrite
	-l,-s	Create (sym)links
	-T	Refuse to copy if DEST is a directory
	-t DIR	Copy all SOURCEs into DIR
	-u	Copy only newer files`,
	"dd": busyboxHeader + `

Usage: dd [if=FILE] [of=FILE] [ibs=N obs=N/bs=N] [count=N] [skip=N] [seek=N]
	[conv=notrunc|noerror|sync|fsync]
	[iflag=skip_bytes|count_bytes|fullblock|direct] [oflag=seek_bytes|append|direct]

Copy a file with converting and formatting

	if=FILE		Read from FILE instead of stdin
	of=FILE		Write to FILE instead of stdout
	bs=N		Read and write N bytes at a time
	ibs=N		Read N bytes at a time
	obs=N		Write N bytes at a time
	count=N		Copy only N input blocks
	skip=N		Skip N input blocks
	seek=N		Skip N output blocks
	conv=notrunc	Don't truncate output file
	conv=noerror	Continue after read errors
	conv=sync	Pad blocks with zeros
	conv=fsync	Physically write data out before finishing
	conv=swab	Swap every pair of bytes
	iflag=skip_bytes	skip=N is in bytes
	iflag=count_bytes	count=N is in bytes
	oflag=seek_bytes	seek=N is in bytes
	iflag=direct	O_DIRECT input
	oflag=direct	O_DIRECT output
	iflag=fullblock	Read full blocks
	oflag=append	Open output in append mode

N may be suffixed by c (1), w (2), b (512), kB (1000), k (1024), MB, M, GB, G`,
	"uname": busyboxHeader + `

Usage: uname [-amnrspvio]

Print system information

        -a      Print all
        -m      Machine (hardware) type
        -n      Hostname
        -r      Kernel release
        -s      Kernel name (default)
        -p      Processor type
        -v      Kernel version
        -i      Hardware platform
        -o      OS name
`,
	"rm": busyboxHeader + `

Usage: rm [-irf] FILE...

Remove (unlink) FILEs

        -i      Always prompt before removing
        -f      Never prompt
        -R,-r   Recurse
`,
	"wc": busyboxHeader + `

Usage: wc [-clwL] [FILE]...

Count lines, words, and bytes for FILEs (or stdin)

	-c	Count bytes
	-l	Count newlines
	-w	Count words
	-L	Print longest line length`,
}
