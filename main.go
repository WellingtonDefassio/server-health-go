package main

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type Server struct {
	name     string
	url      string
	time     float64
	status   string
	dateFail string
}

func main() {
	file, downTime := openFiles(os.Args[1], os.Args[2])
	defer file.Close()
	defer downTime.Close()
	serverList := createServerList(file)
	downServers := checkServerHealth(serverList)
	genereteDowntime(downTime, downServers)

}

func genereteDowntime(downTime *os.File, servers []Server) {
	writer := csv.NewWriter(downTime)
	for _, server := range servers {
		line := []string{server.name, server.url, server.dateFail, fmt.Sprintf("%f", server.time), fmt.Sprintf("%s", server.status)}
		writer.Write(line)
	}
	writer.Flush()
}

func openFiles(pServerFile string, pDownTimeFile string) (*os.File, *os.File) {
	file, err := os.OpenFile(pServerFile, os.O_RDONLY, 0666)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	downTime, err := os.OpenFile(pDownTimeFile, os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return file, downTime
}

func checkServerHealth(serverList []Server) []Server {
	var downServers []Server
	now := time.Now()
	for _, server := range serverList {
		initTime := time.Now()
		get, err := http.Get(server.url)
		if err != nil {
			fmt.Printf("Server %s is down [%s]\n", server.name, err.Error())
			server.status = "Error"
			server.dateFail = now.Format("02/01/2006 15:04:05")
			downServers = append(downServers, server)
			continue
		}
		server.status = get.Status
		if !strings.Contains(server.status, "200") {
			server.dateFail = now.Format("02/01/2006 15:04:05")
			downServers = append(downServers, server)
		}
		server.time = time.Since(initTime).Seconds()
		fmt.Printf("The Server: [%s], Get Response in: [%f], with Status: [%s] \n", server.name, server.time, server.status)
	}
	return downServers
}

func createServerList(serverList *os.File) []Server {
	csvReader := csv.NewReader(serverList)
	data, err := csvReader.ReadAll()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var servers []Server
	for i, v := range data {
		if i > 0 {
			server := Server{
				name: v[0],
				url:  v[1],
			}
			servers = append(servers, server)
		}
	}
	return servers
}
