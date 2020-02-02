package main

import (
	"flag"
	"hahajing/com"
	"hahajing/kad"
	"hahajing/web"
	"os"
	"path/filepath"
)

var kadInstance kad.Kad
var webInstance web.Web


func main() {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	config := com.Config{}
	flag.IntVar(&config.UDPPort,"udp-listen-port",2500,"UDP Listen Port")
	flag.IntVar(&config.ExternalUDPPort,"public-udp-listen-port",2500,"Public UDP Listen Port")
	flag.IntVar(&config.TCPPort,"tcp-listen-port",2501,"TCP Listen Port")
	flag.StringVar(&config.NodeDatPath,"nodesdat-path",dir+"/nodes.dat","nodes.dat path")
	flag.IntVar(&config.SearchTimeWithoutResults,"timeout-noresults",8,"Time to finish search after no more results")
	flag.IntVar(&config.SearchExpires,"search-cache-timeout",60,"Time to cache searches")
	flag.IntVar(&config.MaxContacts,"contacts",5000,"Max number of contacts")
	flag.Parse()

	kadInstance.Start(&config)
	webInstance.Start(kadInstance.SearchReqCh,&config)
}
