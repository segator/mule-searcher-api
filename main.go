package main

import (
	"flag"
	"fmt"
	"hahajing/com"
	"hahajing/download"
	"hahajing/kad"
	"hahajing/publish"
	"hahajing/web"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var kadInstance kad.Kad
var webInstance web.Web
var publisherInstance publish.Publisher
type arrayFlags []string
func (i *arrayFlags) String() string {
	return ""
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {
	config := com.Config{}
	flag.BoolVar(&config.EnableSearcher,"search-enable",true,"Enable searcher? by default true")
	flag.StringVar(&config.HTTPUser,"http-user","admin","http auth user")
	flag.StringVar(&config.HTTPPassword,"http-password","admin","http auth password")
	flag.IntVar(&config.WEBListenPort,"web-listen-port",80,"Web Listen Port")
	flag.IntVar(&config.UDPPort,"udp-listen-port",2500,"UDP Listen Port")
	flag.IntVar(&config.ExternalUDPPort,"public-udp-listen-port",2500,"Public UDP Listen Port")
	flag.IntVar(&config.NumberOfSocket,"NumberOfUDPSockets",10,"Number of UDP Sockets")
	flag.IntVar(&config.TCPPort,"tcp-listen-port",2501,"TCP Listen Port")
	flag.StringVar(&config.NodeDatPath,"nodesdat-path","http://upd.emule-security.org/nodes.dat","nodes.dat path can be http:// or file://")
	flag.IntVar(&config.SearchTimeWithoutResults,"timeout-noresults",8,"Time to finish search after no more results")
	flag.IntVar(&config.SearchExpires,"search-cache-timeout",60,"Time to cache searches")
	flag.IntVar(&config.MaxContacts,"contacts",5000,"Max number of contacts")

	//Send links to download on emule or synology
	var downloaders arrayFlags
	flag.Var(&downloaders, "downloader", "emule:http://password@localhost:4711 or synology:http://user:password@hello.synology.me:5000/downloadpath or amule:tcp://password@localhost:4712 repetable param for multiple downloaders")



	flag.StringVar(&config.DownloadPath,"download-path","/downloads","Path where downloads are saved for emule/synology")
	flag.StringVar(&config.PublishSSHHost,"publish-ssh-host","localhost","SSH Host to publish new downloads")
	flag.IntVar(&config.PublishSSHPort,"publish-ssh-port",22,"SSH Port of the publisher ssh host")
	flag.StringVar(&config.PublishSSHUsername,"publish-ssh-username","root","SSH Username of the publisher ssh host")
	flag.StringVar(&config.PublishSSHPassword,"publish-ssh-password","","SSH Password of the publisher ssh host")
	flag.StringVar(&config.PublishSSHPath,"publish-ssh-path","","SSH Path of the publisher ssh host")
	flag.IntVar(&config.PublishScanTime,"publish-scan-time",60,"Scan Download folder every x minutes")
	flag.IntVar(&config.PublishMinimumTime,"publish-minimum-push-time",60,"minimum life time of a file to be selected as publishable in minutes")
	flag.Parse()


	pattern := regexp.MustCompile(`^(emule|synology|amule):(https?|tcp):\/\/(.+)@(.+):([0-9]+)\/?(.*)?$`)
	multiDownloader := download.MultiDownloader{}
	for _,downloaderString := range downloaders {
		if !pattern.MatchString(downloaderString) {
			com.HhjLog.Errorf("Invalid downloader format: %s", downloaderString)
			os.Exit(1)
		}
		var downloader download.Downloader
		switch pattern.FindStringSubmatch(downloaderString)[1]{
		case "emule":
			emuleURL := fmt.Sprintf("%s://%s:%s",pattern.FindStringSubmatch(downloaderString)[2],pattern.FindStringSubmatch(downloaderString)[4],pattern.FindStringSubmatch(downloaderString)[5])
			downloader = download.EmuleDownloader{Password:pattern.FindStringSubmatch(downloaderString)[3], EmuleWebURL:emuleURL}
		case "synology":
			synologyURL := fmt.Sprintf("%s://%s:%s",pattern.FindStringSubmatch(downloaderString)[2],pattern.FindStringSubmatch(downloaderString)[4],pattern.FindStringSubmatch(downloaderString)[5])
			userPass := strings.Split(pattern.FindStringSubmatch(downloaderString)[3],":")
			downloader = download.SynologyMuleDownloader{SynologyPassword:userPass[1],
				SynologyUser:userPass[0],
				SynologyURL:synologyURL,
				SynologyDestionation: pattern.FindStringSubmatch(downloaderString)[6],
			}
		case "amule":
			amulePort, err := strconv.Atoi(pattern.FindStringSubmatch(downloaderString)[5])
			if err != nil {
				com.HhjLog.Errorf("Invalid Amule Port format: %s", pattern.FindStringSubmatch(downloaderString)[5])
				os.Exit(1)
			}
			downloader = download.AmuleDownloader{AmuleHost:pattern.FindStringSubmatch(downloaderString)[4], AmulePort:amulePort,AmulePassword:pattern.FindStringSubmatch(downloaderString)[3],}
		}
		multiDownloader.DownloaderList = append(multiDownloader.DownloaderList,downloader)
	}


	if config.EnableSearcher {
		kadInstance.Start(&config)
		webInstance.Start(kadInstance.SearchReqCh,&config,multiDownloader)
	}

	if config.PublishSSHPath != "" {
		publisher := publish.PublisherSSHConfig{
			Config:             publish.PublisherConfig{
				DownloadPath:config.DownloadPath,
				ValidUploadableFormats: []string{"mkv","mp4","avi"},
			},
			ScanTime: time.Duration(config.PublishScanTime) * time.Minute,
			PublishSSHHost:     config.PublishSSHHost,
			PublishSSHUsername:  config.PublishSSHUsername,
			PublishSSHPassword: config.PublishSSHPassword,
			PublishSSHPath:     config.PublishSSHPath,
			PublishSSHPort:     config.PublishSSHPort,
			PublishMinimumTime: time.Duration(config.PublishMinimumTime) * time.Minute,
		}
		publisher.Start()
	}


}
