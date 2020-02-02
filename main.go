package main

import (
	"hahajing/kad"
	"hahajing/web"
)

var kadInstance kad.Kad
var webInstance web.Web


func main() {
	kadInstance.Start()
	webInstance.Start(kadInstance.SearchReqCh)
}
