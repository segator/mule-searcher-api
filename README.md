# mule-searcher-api
This application allow to find e2dk on KAD Network via API, thanks to https://github.com/moyuanz/hahajing
I created this application to allow sickrage and forks to support downloading of e2dk.
![Build](https://github.com/segator/mule-searcher-api/workflows/Build/badge.svg)

This application works with jackett, check my fork of jackett to support mule-searcher-api
https://github.com/segator/Jackett/tree/feature/e2dkkad

Sickrage --> Jackett --> mule-searcher-api

When sickrage nows what he wants to download download a fake torrent generated by mule-seracher-api (The e2dk is as announce inside the fake torrent)

Then you need to configure sickrage to connect to a qbitorrent to send the torrents.
mule-searcher-api implements qbittorrent protocol, and then capture the torrents and send it to an emule

## Build

## RUN
Can be done using the binary or the docker image 

### Command Line
```bash
  -NumberOfUDPSockets int
    	Number of UDP Sockets (default 10)
  -contacts int
    	Max number of contacts (default 5000)
  -emule-password string
    	admin (default "admin")
  -emule-url string
    	Emule URL (default "http://localhost:4711")
  -nodesdat-path string
    	nodes.dat path can be http:// or file:// (default "http://upd.emule-security.org/nodes.dat")
  -public-udp-listen-port int
    	Public UDP Listen Port (default 2500)
  -search-cache-timeout int
    	Time to cache searches (default 60)
  -tcp-listen-port int
    	TCP Listen Port (default 2501)
  -timeout-noresults int
    	Time to finish search after no more results (default 8)
  -udp-listen-port int
    	UDP Listen Port (default 2500)
  -web-listen-port int
    	Web Listen Port (default 80)
```