package main

import (
	"fmt"
	"github.com/xinxindu/logger"
	"time"
)

func main() {
	log, err := logger.InitLogger("M", 1024, logger.LevelInfo, "c:/Users/dxx/go/src/github.com/xinxindu/logger",
		"mylog")
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		log.Infof("%s  %s", "aaa", "bbbb")
		time.Sleep(5 * time.Second)
	}

	log.Close()
}