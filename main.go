package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

var flags struct {
	configPath *string
	noUpdate   *bool
}

func main() {
	flags.configPath = flag.String("config", "config.toml", "Path to config file")
	flags.noUpdate = flag.Bool("no-update", false, "Don't update mirrors automatically")
	flag.Parse()

	if flags.configPath == nil {
		log.Fatal("The -config flag is mandatory, an example config is available at https://github.com/beefsack/git-mirror/blob/master/example-config.toml")
	}

	cfg, repos, err := parseConfig(*flags.configPath)
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
				if *flags.noUpdate {
					break
				}
				time.Sleep(r.Interval.Duration)
			}
		}(r)
	}

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		parts := strings.SplitN(request.URL.Path, "/", 5)
		combined := strings.Join(parts[1:4], "/")
		_, statErr := os.Stat(path.Join(cfg.BasePath, combined))
		if os.IsNotExist(statErr) {
			log.Printf("Repository not in cache: %s", combined)
			r := repo{
				Name:     combined,
				Origin:   fmt.Sprintf("https://%s", combined),
				Interval: duration{Duration: cfg.Interval.Duration},
			}
			err = mirror(cfg, r)
			if err != nil {
				log.Printf("Mirror-on-demand error: %s", err)
			}
		}

		// file exists (now), let the builtin handler serve it
		http.FileServer(http.Dir(cfg.BasePath)).ServeHTTP(writer, request)
	})

	log.Printf("starting web server on %s", cfg.ListenAddr)
	if err := http.ListenAndServe(cfg.ListenAddr, nil); err != nil {
		log.Fatalf("failed to start server, %s", err)
	}
}
