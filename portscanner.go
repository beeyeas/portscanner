package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
)

type PortScannerConfig struct {
	Portrange string
	Ipaddress string
	Protocol  string
}

type PortScannerResult struct {
	portScannerResult portScannerResultMap
	running           int
	timeOut           int
}

type portScannerResultMap map[string]bool

var portScannerTuple PortScannerResult

func main() {
	log.SetLevel(log.DebugLevel)
	log.Infoln("*******************************************")
	log.Infoln("Port Scanner")
	log.Infoln("*******************************************")

	t := time.Now()
	defer func() {
		if e := recover(); e != nil {
			log.Debugln(e)
		}
	}()

	log.Debugln(loadConfig("config.properties"))

	log.Debugln("Parsed input data ", len(portScannerTuple.portScannerResult))
	CheckPort(&portScannerTuple)

	for key, value := range portScannerTuple.portScannerResult {
		if value {
			log.Debugln("Port Scanner Result", key, " port is open :", value)
		}
	}

	log.Debugln("Total time taken %s to scan %d ports", time.Since(t), len(portScannerTuple.portScannerResult))
}

func CheckPort(portScannerTuple *PortScannerResult) {
	for record := range portScannerTuple.portScannerResult {
		for portScannerTuple.running >= 2000 {
			log.Debugln("Maximum threads spawned", portScannerTuple.running, " waiting ...", 1*time.Second)
			time.Sleep(1 * time.Second)
		}
		r := strings.Split(record, ":")
		port, _ := strconv.Atoi(r[1])
		portScannerTuple.running++
		go check(portScannerTuple, r[0], uint16(port))
	}

	for portScannerTuple.running != 0 {
		time.Sleep(1 * time.Second)
	}
}

func check(portScannerTuple *PortScannerResult, ip string, port uint16) {
	connection, err := net.DialTimeout("tcp", ip+":"+fmt.Sprintf("%d", port), time.Duration(portScannerTuple.timeOut)*time.Second)
	if err == nil {
		portScannerTuple.portScannerResult[fmt.Sprintf("%s:%d", ip, port)] = true
		//log.Debugln(fmt.Sprintf("%s:%d - true", ip, port))
		connection.Close()
	} else {
		portScannerTuple.portScannerResult[fmt.Sprintf("%s:%d", ip, port)] = false
		//log.Debugln(fmt.Sprintf("%s:%d - %s", ip, port, err))
	}
	portScannerTuple.running--
}

func loadConfig(file string) PortScannerConfig {
	var readConfigStruct PortScannerConfig
	if metaData, err := toml.DecodeFile(file, &readConfigStruct); err != nil {
		log.Debugln("Error Occured Reading file", err, metaData)
	}
	ports := strings.Split(readConfigStruct.Portrange, "-")
	p1, err := strconv.Atoi(ports[0])
	if err != nil {
		log.Errorln(err)
	}
	log.Debugln("p1", p1)
	p2, err := strconv.Atoi(ports[1])
	if err != nil {
		log.Errorln(err)
	}
	log.Debugln("p2", p2)
	portScannerTuple.portScannerResult = make(portScannerResultMap)
	for port := p1; port <= p2; port++ {
		portScannerTuple.timeOut = 5
		portScannerTuple.portScannerResult[readConfigStruct.Ipaddress+fmt.Sprintf(":%d", port)] = false
	}
	return readConfigStruct
}
