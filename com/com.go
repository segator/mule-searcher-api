package com

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"fmt"
	"github.com/op/go-logging"
	"io"
	"net/http"
	"os"
	"strings"
)
type Config struct {
	//Searcher parameters
	EnableSearcher           string
	WEBListenPort            int
	HTTPUser                 string
	HTTPPassword             string
	UDPPort                  int
	NumberOfSocket           int
	ExternalUDPPort          int
	TCPPort                  int
	MaxContacts              int
	SearchTimeWithoutResults int
	SearchExpires            int
	NodeDatPath              string

	//Publisher parameters
	DownloadPath             string
	DownloadPathCompleted    string
	PublishSSHHost           string
	PublishSSHUsername       string
	PublishSSHPassword       string
	PublishSSHPath           string
	PublishSSHPathTV         string
	PublishSSHPathMovies     string
	PublishSSHPort           int
	PublishScanTime          int
	PublishMinimumTime       int
	Verbosity                string

}


// HhjLog is HHJ system log
var HhjLog = logging.MustGetLogger("hhj")
var logformat = logging.MustStringFormatter(
	`%{color}%{time:2006-01-02 15:04:05.000} ▶ %{level:.4s} %{color:reset} %{message}`,
)

func LoggerInit(module string) error {
	backend := logging.NewLogBackend(os.Stdout, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, logformat)

	backendLeveled := logging.AddModuleLevel(backendFormatter)
	level,err := logging.LogLevel(module)
	if err!=nil {
		return err
	}
	backendLeveled.SetLevel(level,"")

	// Set the backends to be used.
	logging.SetBackend(backendLeveled)
	return nil
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
	keys := InnerSplit2Keywords(s, ignore)

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
	return InnerSplit2Keywords(s, "")
}

// innerSplit2Keywords is split to slice of keyword by seperators.
// @ignore: which chars not think as seperator.
func InnerSplit2Keywords(s string, ignore string) []string {
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

func SHA1String(string string) string {
	h := sha1.New()
	h.Write([]byte(string))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x",bs)
}