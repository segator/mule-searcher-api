package main

import (
	"flag"
	"hahajing/com"
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
	flag.IntVar(&config.TCPPort,"tcp-listen-port",2501,"TCP Listen Port")
	flag.StringVar(&config.NodeDatPath,"nodesdat-path","http://upd.emule-security.org/nodes.dat","nodes.dat path can be http:// or file://")
	flag.IntVar(&config.SearchTimeWithoutResults,"timeout-noresults",8,"Time to finish search after no more results")
	flag.IntVar(&config.SearchExpires,"search-cache-timeout",60,"Time to cache searches")
	flag.IntVar(&config.MaxContacts,"contacts",5000,"Max number of contacts")
	flag.StringVar(&config.EMuleURL,"emule-url","http://localhost:4711","Emule URL")
	flag.StringVar(&config.EMULEWebPassword,"emule-password","admin","admin")


	flag.Parse()

	kadInstance.Start(&config)
	webInstance.Start(kadInstance.SearchReqCh,&config)
}
