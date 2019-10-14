package main

import (
	common "acous/commom"
	"acous/web/conf"
	"acous/web/service"
	"flag"
	"log"
	"os"
)

var (
	confFile = flag.String("conf", "conf.yml", "The configure file")
	logFile  = flag.String("log", "stdio", "The log file")
)

func main() {
	flag.Parse()
	if *logFile != `stdio` {
		f, err := os.OpenFile(*logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Fatal(err)
		}

		defer func() {
			err = f.Close()
			if err != nil {
				log.Println(err.Error())
			}
		}()

		log.SetOutput(f)
	}

	c := &conf.Conf{}
	c.GetConfOrDie(*confFile)

	//fmt.Println(c)
	//c.UpdateTau("1", 8.88888)
	//fmt.Println(c)
	common.InitWebClient()
	service.StartWebService(c)
}
