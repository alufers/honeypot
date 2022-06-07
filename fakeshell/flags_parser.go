package fakeshell

import (
	"fmt"
	"strconv"
	"strings"
)

//ParseBeginningShortFlags parses flags like "echo -e fdffdfd -C" as "e": true
func ParseBeginningShortFlags(args []string) (flags map[string]bool, rest []string) {
	flags = map[string]bool{}
	rest = []string{}
	for idx, arg := range args {
		if len(arg) > 1 && arg[0] == '-' {
			for _, c := range arg[1:] {
				flags[string(c)] = true
			}
		} else {
			rest = args[idx:]
			break
		}
	}
	return
}

func ParseBeginningShortFlagsValidated(args []string, validFlags string) (flags map[string]bool, rest []string, err error) {
	flags, rest = ParseBeginningShortFlags(args)
	for flag := range flags {
		if !strings.Contains(validFlags, flag) {
			err = fmt.Errorf("invalid flag: %s", flag)
			return
		}
	}
	return
}

//ParseFlagsDDStyle parses flags like "dd if=/dev/zero of=/dev/null bs=1M count=1" as "if": "/dev/zero", "of": "/dev/null", "bs": "1M", "count": "1"
func ParseFlagsDDStyle(args []string) (flags map[string]string, badFlags bool) {
	flags = map[string]string{}

	for _, arg := range args {
		segs := strings.SplitN(arg, "=", 2)
		if len(segs) == 2 {
			flags[segs[0]] = segs[1]
		} else {
			badFlags = true
			flags[segs[0]] = ""
		}
	}
	return
}

//ParseFileSize parses file size for dd
func ParseFileSize(rawSize string) (int64, error) {
	prefixes := map[string]int64{
		"c":  1,
		"w":  2,
		"b":  512,
		"kB": 1000,
		"k":  1024,
		"MB": 1000 * 1000,
		"M":  1024 * 1024,
		"GB": 1000 * 1000 * 1000,
		"G":  1024 * 1024 * 1024,
	}
	multiplier := int64(1)
	for prefix, size := range prefixes {
		if strings.HasPrefix(rawSize, prefix) {
			rawSize = strings.TrimPrefix(rawSize, prefix)
			multiplier = size

		}
	}

	parsed, err := strconv.ParseInt(rawSize, 10, 64)
	if err != nil || parsed < 0 {
		return 0, fmt.Errorf("invalid number '%s'", rawSize)
	}
	return parsed * multiplier, nil

}
