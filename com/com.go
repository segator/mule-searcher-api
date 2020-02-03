package com

import (
	"bufio"
	"bytes"
	"github.com/op/go-logging"
	"io"
	"net/http"
	"os"
	"strings"
)
type Config struct {
	UDPPort  int
	NumberOfSocket int
	ExternalUDPPort int
	TCPPort int
	MaxContacts int
	SearchTimeWithoutResults int
	SearchExpires int
	NodeDatPath string
	EmuleDownloader string
	EMuleURL string
	EMULEWebPassword string
	SynologyUsername string
	SynologyPassword string
	SynologyURL string
	SynologyDestionation string
	WEBListenPort int
}


// HhjLog is HHJ system log
var HhjLog = logging.MustGetLogger("hhj")
var logformat = logging.MustStringFormatter(
	`%{color}%{time:2006-01-02 15:04:05.000} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)

func init() {
	backend := logging.NewLogBackend(os.Stdout, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, logformat)

	backendLeveled := logging.AddModuleLevel(backend)
	backendLeveled.SetLevel(logging.CRITICAL, "")

	// Set the backends to be used.
	logging.SetBackend(backendLeveled, backendFormatter)
}

// GetConfigPath x
func GetConfigPath() string {
	path := os.Args[2]

	i := strings.LastIndex(path, "\\")
	if i == -1 {
		i = strings.LastIndex(path, "/")
	}

	if i == -1 {
		HhjLog.Fatalf("Config path error: %s", path)
	}

	path = string(path[0:i])

	return path
}


// Split2PrimaryKeywords is split to slice of primary keyword by seperators.
// And keyword containing specific char not thinking as primary keyword.
func Split2PrimaryKeywords(s string) []string {
	ignore := "'’"
	keys := innerSplit2Keywords(s, ignore)

	var newKeys []string
	for _, key := range keys {
		valid := true
		for _, c := range ignore {
			if strings.Index(key, string(c)) != -1 {
				valid = false
				break
			}
		}

		if valid {
			newKeys = append(newKeys, key)
		}
	}

	return newKeys
}

// Split2Keywords is split to slice of keyword by seperators.
func Split2Keywords(s string) []string {
	return innerSplit2Keywords(s, "")
}

// innerSplit2Keywords is split to slice of keyword by seperators.
// @ignore: which chars not think as seperator.
func innerSplit2Keywords(s string, ignore string) []string {
	sep := `·!/\*?<>|-_:,.;'"()[]‘’“”；、：，。？！` + "\t"
	for _, c := range ignore {
		sep = strings.Replace(sep, string(c), "", -1)
	}

	for _, c := range sep {
		s = strings.Replace(s, string(c), " ", -1)
	}
//ISAAC
	s = strings.ToLower(s)

	var newKeys []string
	for _, key := range strings.Split(s, " ") {
		if key != "" {
			newKeys = append(newKeys, key)
		}
	}

	return newKeys
}

func DownloadFile(url string) (*bytes.Buffer,error) {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return nil,err
	}
	defer resp.Body.Close()
	//size,_ := strconv.Atoi(resp.Header.Get("Content-Length"))
	var b bytes.Buffer
	buf := bufio.NewWriter(&b)
	// Write the body to file
	io.Copy(buf, resp.Body)
	return &b,nil
}