package main

import (
	"flag"
	"log"
	"time"
)

func main() {
	interval := flag.Duration("interval", time.Second*15, "interval between keep awake calls")
	awakeduration := flag.Duration("howlong", time.Minute*10, "how long to keep awake")
	flag.Parse()

	awakeuntil := time.Now().Add(*awakeduration)
	log.Printf("keep awake until %+v", awakeuntil)

	for awakeuntil.After(time.Now()) {
		log.Printf("keep awake for %v", time.Now().Sub(awakeuntil))
		keep_awake()
		time.Sleep(*interval)
	}
}
