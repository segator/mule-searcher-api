package main

import (
	"flag"
	"hahajing/com"
	"hahajing/download"
	"hahajing/kad"
	"hahajing/publish"
	"hahajing/web"
	"time"
)

var kadInstance kad.Kad
var webInstance web.Web
var publisherInstance publish.Publisher


func main() {
	config := com.Config{}

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
	flag.StringVar(&config.EmuleDownloader,"emule-type","emule","Type of Emule Downloader service (emule,synology,amulecmd)")
	//Emule params
	flag.StringVar(&config.EMuleURL,"emule-url","http://localhost:4711","Emule URL")
	flag.StringVar(&config.EMULEWebPassword,"emule-password","admin","admin")
	//Synology Params
	flag.StringVar(&config.SynologyUsername,"synology-username","","Synology username")
	flag.StringVar(&config.SynologyPassword,"synology-password","","Synology password")
	flag.StringVar(&config.SynologyURL,"synology-url","http://192.168.1.20:5000","Synology URL")
	flag.StringVar(&config.SynologyDestionation,"synology-download-path","","Synology download destination path")

	flag.StringVar(&config.AmuleHost,"amule-host","localhost","Amule Daemon Host")
	flag.IntVar(&config.AmulePort,"amule-port",4712,"Amule Daemon Port")
	flag.StringVar(&config.AmulePassword,"amule-password","","Amule Daemon Password")

	flag.StringVar(&config.DownloadPath,"download-path","/downloads","Path where downloads are saved for emule/synology")
	flag.StringVar(&config.PublishSSHHost,"publish-ssh-host","localhost","SSH Host to publish new downloads")
	flag.IntVar(&config.PublishSSHPort,"publish-ssh-port",22,"SSH Port of the publisher ssh host")
	flag.StringVar(&config.PublishSSHUsername,"publish-ssh-username","root","SSH Username of the publisher ssh host")
	flag.StringVar(&config.PublishSSHPassword,"publish-ssh-password","","SSH Password of the publisher ssh host")
	flag.StringVar(&config.PublishSSHPath,"publish-ssh-path","","SSH Path of the publisher ssh host")
	flag.IntVar(&config.PublishScanTime,"publish-scan-time",60,"Scan Download folder every x minutes")
	flag.IntVar(&config.PublishMinimumTime,"publish-minimum-push-time",60,"minimum life time of a file to be selected as publishable in minutes")


	flag.Parse()
	var downloader download.Downloader
	switch config.EmuleDownloader{
	case "emule":
		downloader = download.EmuleDownloader{Password:config.EMULEWebPassword, EmuleWebURL:config.EMuleURL}
	case "synology":
		downloader = download.SynologyMuleDownloader{SynologyPassword:config.SynologyPassword,
			SynologyUser:config.SynologyUsername,
			SynologyURL:config.SynologyURL,
			SynologyDestionation: config.SynologyDestionation,
		}
	case "amule":
		downloader = download.AmuleDownloader{AmuleHost:config.AmuleHost, AmulePort:config.AmulePort,AmulePassword:config.AmulePassword,}
	}
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
	kadInstance.Start(&config)
	publisher.Start()
	webInstance.Start(kadInstance.SearchReqCh,&config,downloader)

}
