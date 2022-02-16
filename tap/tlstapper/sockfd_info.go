package tlstapper

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-errors/errors"
)

var socketInodeRegex = regexp.MustCompile(`socket:\[(\d+)\]`)

const (
	SRC_ADDRESS_FILED_INDEX = 1
	DST_ADDRESS_FILED_INDEX = 2
	INODE_FILED_INDEX       = 9
)

func getAddressBySockfd(procfs string, pid uint32, fd uint32, src bool) (net.IP, uint16, error) {
	inode, err := getSocketInode(procfs, pid, fd)

	if err != nil {
		return nil, 0, err
	}

	tcppath := fmt.Sprintf("%s/%d/net/tcp", procfs, pid)
	tcp, err := ioutil.ReadFile(tcppath)

	if err != nil {
		return nil, 0, errors.Wrap(err, 0)
	}

	for _, line := range strings.Split(string(tcp), "\n") {
		parts := strings.Fields(line)

		if len(parts) < 10 {
			continue
		}

		if inode == parts[INODE_FILED_INDEX] {
			if src {
				return parseHexAddress(parts[SRC_ADDRESS_FILED_INDEX])
			} else {
				return parseHexAddress(parts[DST_ADDRESS_FILED_INDEX])
			}
		}
	}

	return nil, 0, errors.Errorf("address not found [pid: %d] [sockfd: %d] [inode: %s]", pid, fd, inode)
}

func getSocketInode(procfs string, pid uint32, fd uint32) (string, error) {
	fdlinkPath := fmt.Sprintf("%s/%d/fd/%d", procfs, pid, fd)
	fdlink, err := os.Readlink(fdlinkPath)

	if err != nil {
		return "", errors.Wrap(err, 0)
	}

	tokens := socketInodeRegex.FindStringSubmatch(fdlink)

	if tokens == nil || len(tokens) < 1 {
		return "", errors.Errorf("socket inode not found [pid: %d] [sockfd: %d] [link: %s]", pid, fd, fdlink)
	}

	return tokens[1], nil
}

// Format looks like 0100007F:50 for 127.0.0.1:80
//
func parseHexAddress(addr string) (net.IP, uint16, error) {
	addrParts := strings.Split(addr, ":")

	port, err := strconv.ParseUint(addrParts[1], 16, 16)

	if err != nil {
		return nil, 0, errors.Wrap(err, 0)
	}

	ip, err := strconv.ParseUint(addrParts[0], 16, 32)

	if err != nil {
		return nil, 0, errors.Wrap(err, 0)
	}

	return net.IP{uint8(ip), uint8(ip >> 8), uint8(ip >> 16), uint8(ip >> 24)}, uint16(port), nil
}
