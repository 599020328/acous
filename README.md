The golang implementation of acous   

## Introduction
----
dir **commom** include the web request of this algorithm.   
dir **web** include the configuration, api, handler, implement of this algorithm.

## Usage
----
Use `go build main.go` generate main(main.exe in windows), put main(main.exe in windows) and conf.yml in a same dir, modify conf.yml to the real world set, and run `sudo ./main -log %log file path%` in each node, request the source node with json string like.
```
{
	"source_url" : "http://127.0.0.1:8000",
	"work_chain": "2-3-4-5-6-7"
}
```
source_url is the source of the ant, work_chain is all the dest you want to find.   
Or use source code in each node, and `go run main.go -log %log file path%`, and the same with above.
