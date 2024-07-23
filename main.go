package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	configPath := flag.String("config", "config.toml", "Path to config file")
	noUpdate := flag.Bool("no-update", false, "Don't update mirrors automatically")
	flag.Parse()

	if configPath == nil {
		log.Fatal("The -config flag is mandatory, an example config is available at https://github.com/beefsack/git-mirror/blob/master/example-config.toml")
	}

	cfg, repos, err := parseConfig(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	if err = os.MkdirAll(cfg.BasePath, 0755); err != nil {
		log.Fatalf("failed to create %s, %s", cfg.BasePath, err)
	}

	// Run background threads to keep mirrors up to date.
	for _, r := range repos {
		go func(r repo) {
			for {
				log.Printf("updating %s", r.Name)
				if err := mirror(cfg, r); err != nil {
					log.Printf("error updating %s, %s", r.Name, err)
				} else {
					log.Printf("updated %s", r.Name)
				}
				if *noUpdate {
					break
				}
				time.Sleep(r.Interval.Duration)
			}
		}(r)
	}

	// Run HTTP server to serve mirrors.
	http.Handle("/", http.FileServer(http.Dir(cfg.BasePath)))
	log.Printf("starting web server on %s", cfg.ListenAddr)
	if err := http.ListenAndServe(cfg.ListenAddr, nil); err != nil {
		log.Fatalf("failed to start server, %s", err)
	}
}
