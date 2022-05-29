package main

import (
	"log"

	"github.com/ip2location/ip2location-go/v9"
)

var ipDB *ip2location.DB

func init() {
	var err error
	ipDB, err = ip2location.OpenDB(getEnv("IP2LOCATION_DB", "IP2LOCATION-LITE-DB11.IPV6.BIN"))
	if err != nil {
		panic(err)
	}
}
func AddIpLocationDataToAttack(a *Attack) {
	if a.SourceIP == "" {
		return
	}
	ip, err := ipDB.Get_all(a.SourceIP)
	if err != nil {
		log.Printf("failed to get ip location: %v", err)
		return
	}
	a.Country = ip.Country_long
	a.CountryCode = ip.Country_short
	a.Location = ip.City
	a.ISP = ip.Isp
}
