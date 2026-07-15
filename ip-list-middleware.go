package xf

import (
	"github.com/gofiber/fiber/v3"
	"strings"
	"net"
	"fmt"
)

// @param ips ip list, format of item : x.x.x.x or x.x.x.x/n
func CreateIpWhitelistFilter(ips []string) fiber.Handler {
	ipList, err := InitIpList(ips)
	if err != nil {
		panic(err)
	}

	return func(c fiber.Ctx) error {
		if ipList == nil {
			return c.Next()
		}

		remoteAddr := getRemoteAddr(c)
		// fmt.Printf("remoteAddr of getRemoteAddr: %s\n", remoteAddr)
		if len(remoteAddr) == 0 {
			return nil // filtered
		}
		if net.ParseIP(remoteAddr) == nil {
			return nil // 
		}
		if !ipList.Contains(remoteAddr) {
			return nil
		}
		return c.Next()
	}
}

func getRemoteAddr(c fiber.Ctx) string {
	xForwardedFor := c.IPs()
	if len(xForwardedFor) == 0 {
		return c.IP()
	}
	end := len(xForwardedFor) - 1
	return xForwardedFor[end]
}

// --- ip list ----- 
type ips struct {
	isRange  bool
	ip       net.IP
	ipNet   *net.IPNet
}

type IpList struct {
	list []*ips
}

// @param list  item format: x.x.x.x or x.x.x.x/n
func InitIpList(list []string) (ipList *IpList, err error) {
	if len(list) == 0 {
		return
	}

	res := make([]*ips, len(list))
	count := 0

	for _, ip := range list {
		if len(ip) == 0 {
			continue
		}

		if strings.IndexByte(ip, '/') > 0 {
			_, ipNet, e := net.ParseCIDR(ip)
			if e != nil {
				err = e
				return
			}
			res[count] = &ips{
				isRange: true,
				ipNet: ipNet,
			}
			count += 1
			continue
		}

		netIP := net.ParseIP(ip)
		if netIP == nil {
			err = fmt.Errorf("ip %s is invalid", ip)
			continue
		}
		res[count] = &ips{
			ip: netIP,
		}
		count += 1
	}

	if count == 0 {
		return
	}

	ipList = &IpList{
		list: res[:count],
	}
	return
}

func (ipList *IpList) Contains(ip string) bool {
	netIP := net.ParseIP(ip)
	if netIP == nil {
		return false
	}

	for _, ipl := range ipList.list {
		if ipl.isRange {
			if ipl.ipNet.Contains(netIP) {
				return true
			}
		} else {
			if ipl.ip.Equal(netIP) {
				return true
			}
		}
	}
	return false
}
