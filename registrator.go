package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	dockerevents "github.com/docker/docker/api/types/events"
	dockerapi "github.com/fsouza/go-dockerclient"
	"github.com/gliderlabs/pkg/usage"
	"github.com/simplesurance/registrator/bridge"
)

var Version string

var versionChecker = usage.NewChecker("registrator", Version)

var (
	hostIp           = flag.String("ip", "", "IP for ports mapped to the host")
	internal         = flag.Bool("internal", false, "Use internal ports instead of published ones")
	explicit         = flag.Bool("explicit", false, "Only register containers which have SERVICE_NAME label set")
	useIpFromLabel   = flag.String("useIpFromLabel", "", "Use IP which is stored in a label assigned to the container")
	refreshInterval  = flag.Int("ttl-refresh", 0, "Frequency with which service TTLs are refreshed")
	refreshTtl       = flag.Int("ttl", 0, "TTL for services (default is no expiry)")
	forceTags        = flag.String("tags", "", "Append tags for all registered services")
	resyncInterval   = flag.Int("resync", 0, "Frequency with which services are resynchronized")
	deregister       = flag.String("deregister", "always", "Deregister exited services \"always\" or \"on-success\"")
	deregisterOnStop = flag.Bool("deregister-on-stop", false, "Deregister when container stopped versus once it dies")
	retryAttempts    = flag.Int("retry-attempts", 0, "Max retry attempts to establish a connection with the backend. Use -1 for infinite retries")
	retryInterval    = flag.Int("retry-interval", 2000, "Interval (in millisecond) between retry-attempts.")
	cleanup          = flag.Bool("cleanup", false, "Remove dangling services")
)

func assert(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		versionChecker.PrintVersion()
		os.Exit(0)
	}
	log.Printf("Starting registrator %s ...", Version)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s [options] <registry URI>\n\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NArg() != 1 {
		if flag.NArg() == 0 {
			fmt.Fprint(os.Stderr, "Missing required argument for registry URI.\n\n")
		} else {
			fmt.Fprintln(os.Stderr, "Extra unparsed arguments:")
			fmt.Fprintln(os.Stderr, " ", strings.Join(flag.Args()[1:], " "))
			fmt.Fprint(os.Stderr, "Options should come before the registry URI argument.\n\n")
		}
		flag.Usage()
		os.Exit(2)
	}

	if *hostIp != "" {
		log.Println("Forcing host IP to", *hostIp)
	}

	if (*refreshTtl == 0 && *refreshInterval > 0) || (*refreshTtl > 0 && *refreshInterval == 0) {
		assert(errors.New("-ttl and -ttl-refresh must be specified together or not at all"))
	} else if *refreshTtl > 0 && *refreshTtl <= *refreshInterval {
		assert(errors.New("-ttl must be greater than -ttl-refresh"))
	}

	if *retryInterval <= 0 {
		assert(errors.New("-retry-interval must be greater than 0"))
	}

	dockerHost := os.Getenv("DOCKER_HOST")
	if dockerHost == "" {
		if runtime.GOOS != "windows" {
			os.Setenv("DOCKER_HOST", "unix:///tmp/docker.sock") //nolint: errcheck
		} else {
			os.Setenv("DOCKER_HOST", "npipe:////./pipe/docker_engine") //nolint: errcheck
		}
	}

	docker, err := dockerapi.NewClientFromEnv()
	assert(err)

	if *deregister != "always" && *deregister != "on-success" {
		assert(errors.New("-deregister must be \"always\" or \"on-success\""))
	}

	b, err := bridge.New(docker, flag.Arg(0), bridge.Config{
		HostIp:          *hostIp,
		Internal:        *internal,
		Explicit:        *explicit,
		UseIpFromLabel:  *useIpFromLabel,
		ForceTags:       *forceTags,
		RefreshTtl:      *refreshTtl,
		RefreshInterval: *refreshInterval,
		DeregisterCheck: *deregister,
		Cleanup:         *cleanup,
	})

	assert(err)

	attempt := 0
	for *retryAttempts == -1 || attempt <= *retryAttempts {
		log.Printf("Connecting to backend (%v/%v)", attempt, *retryAttempts)

		err = b.Ping()
		if err == nil {
			break
		}

		if err != nil && attempt == *retryAttempts {
			assert(err)
		}

		time.Sleep(time.Duration(*retryInterval) * time.Millisecond)
		attempt++
	}

	// Start event listener before listing containers to avoid missing anything
	events := make(chan *dockerapi.APIEvents, 128)
	assert(docker.AddEventListener(events))
	log.Println("Listening for Docker events ...")

	b.Sync(false)

	quit := make(chan struct{})

	// Start the TTL refresh timer
	if *refreshInterval > 0 {
		ticker := time.NewTicker(time.Duration(*refreshInterval) * time.Second)
		go func() {
			for {
				select {
				case <-ticker.C:
					b.Refresh()
				case <-quit:
					ticker.Stop()
					return
				}
			}
		}()
	}

	// Start the resync timer if enabled
	if *resyncInterval > 0 {
		resyncTicker := time.NewTicker(time.Duration(*resyncInterval) * time.Second)
		go func() {
			for {
				select {
				case <-resyncTicker.C:
					b.Sync(true)
				case <-quit:
					resyncTicker.Stop()
					return
				}
			}
		}()
	}

	// Process Docker events
	for msg := range events {
		switch dockerevents.Action(msg.Action) {
		case dockerevents.ActionStart, dockerevents.ActionUnPause:
			go b.Add(msg.ID)
		case dockerevents.ActionDie:
			go b.RemoveOnExit(msg.ID)
		case dockerevents.ActionKill:
			if *deregisterOnStop {
				go b.RemoveOnExit(msg.ID)
			}
		}
	}

	close(quit)
	log.Fatal("Docker event loop closed") // todo: reconnect?
}
