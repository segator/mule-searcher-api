package main

import (
	"flag"
	"hahajing/com"
	"hahajing/download"
	"hahajing/kad"
	"hahajing/web"
)

var kadInstance kad.Kad
var webInstance web.Web


func main() {
	config := com.Config{}
	flag.IntVar(&config.WEBListenPort,"web-listen-port",80,"Web Listen Port")
	flag.IntVar(&config.UDPPort,"udp-listen-port",2500,"UDP Listen Port")
	flag.IntVar(&config.ExternalUDPPort,"public-udp-listen-port",2500,"Public UDP Listen Port")
	flag.IntVar(&config.NumberOfSocket,"NumberOfUDPSockets",10,"Number of UDP Sockets")
	flag.IntVar(&config.TCPPort,"tcp-listen-port",2501,"TCP Listen Port")
	flag.StringVar(&config.NodeDatPath,"nodesdat-path","http://upd.emule-security.org/nodes.dat","nodes.dat path can be http:// or file://")
	flag.IntVar(&config.SearchTimeWithoutResults,"timeout-noresults",8,"Time to finish search after no more results")
	flag.IntVar(&config.SearchExpires,"search-cache-timeout",60,"Time to cache searches")
	flag.IntVar(&config.MaxContacts,"contacts",5000,"Max number of contacts")
	flag.StringVar(&config.EmuleDownloader,"emule-type","emule","Type of Emule Downloader service (emule,synology)")
	flag.StringVar(&config.EMuleURL,"emule-url","http://localhost:4711","Emule URL")
	flag.StringVar(&config.EMULEWebPassword,"emule-password","admin","admin")
	flag.StringVar(&config.SynologyUsername,"synology-username","","Synology username")
	flag.StringVar(&config.SynologyPassword,"synology-password","","Synology password")
	flag.StringVar(&config.SynologyURL,"synology-url","http://192.168.1.20:5000","Synology URL")
	flag.StringVar(&config.SynologyDestionation,"synology-destionation","","Synology destination path")
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
	}
	kadInstance.Start(&config)
	webInstance.Start(kadInstance.SearchReqCh,&config,downloader)
}
