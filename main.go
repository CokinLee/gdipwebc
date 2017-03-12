package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ktateish/gdipwebc/gdip"
)

var (
	serviceURL string
	username   string
	password   string
	domainName string
	address    string
	offline    bool
	interval   time.Duration
	debug      bool

	daemon bool // daemon mode (i.e. interval is not zero)
)

func logFatal(format string, args ...interface{}) {
	if daemon {
		log.Fatalf("F: "+format, args...)
	} else {
		fmt.Fprintf(os.Stderr, format, args...)
	}
}

func logError(format string, args ...interface{}) {
	if daemon {
		log.Printf("E: "+format, args...)
	} else {
		fmt.Fprintf(os.Stderr, format, args...)
	}
}

func logInfo(format string, args ...interface{}) {
	if daemon {
		log.Printf("I: "+format, args...)
	}
}

func logdebug(format string, args ...interface{}) {
	if debug {
		log.Printf("D: "+format, args...)
	}
}

func init() {
	flag.StringVar(&serviceURL, "url", "", "URL of a GnuDIP service. required.")
	flag.StringVar(&username, "user", "", "User name of the service. required.")
	flag.StringVar(&password, "password", "", "Password of the service. required.")
	flag.StringVar(&domainName, "domain-name", "", "Domain name to be updated. required.")
	flag.StringVar(&address, "address", "", "IPv4 address for the domain name. If not set, the address seen by the server is used.")
	flag.BoolVar(&offline, "offline", false, "Instruct the service that the domain name is to be offline.")
	flag.DurationVar(&interval, "interval", 0, "Update interval.  Specifying '0' means that the update request is one-shot.")
	flag.BoolVar(&debug, "debug", false, "Enable debug log.")
}

func setFromEnv(f *flag.Flag) {
	envKey := "GDIPWEBC_" + strings.Replace(strings.ToUpper(f.Name), "-", "_", -1)
	envVal := os.Getenv(envKey)
	if envVal != "" {
		err := flag.Set(f.Name, envVal)
		if err != nil {
			logFatal("Cannot set flag %s=%s from environment %s: %v", f.Name, envVal, envKey, err)
		}
	}
}

func validateFlags() (gdip.RequestCode, error) {
	mode := gdip.REGISTER_PASSIVE

	if serviceURL == "" {
		return mode, fmt.Errorf("--url is required")
	}

	if username == "" {
		return mode, fmt.Errorf("--user is required")
	}

	// empty password should be allowed?
	if password == "" {
		return mode, fmt.Errorf("--password is required")
	}

	if domainName == "" {
		return mode, fmt.Errorf("--domain-name is required")
	}

	if address != "" && offline {
		return mode, fmt.Errorf("Cannot set --address with --offline")
	}

	if address != "" {
		mode = gdip.REGISTER
	}

	if offline {
		mode = gdip.OFFLINE
	}

	if interval > 0 {
		daemon = true
	}

	return mode, nil
}

func daemonLoop(g *gdip.Client) {
	for {
		addr, err := g.Update()
		if err != nil {
			logError("Failed to update %s: %v", g.DomainName, err)
		} else {
			logInfo("Successfully update: %s -> %s", g.DomainName, addr)
		}
		<-time.After(interval)
	}
}

func main() {
	flag.VisitAll(setFromEnv)
	flag.Parse()
	if debug {
		gdip.DebugOn()
	}
	mode, err := validateFlags()
	if err != nil {
		if daemon {
			logFatal("Invalid options: %v", err)
		} else {
			fmt.Fprintf(os.Stderr, "Invalid options: %v\n\n", err)
			flag.Usage()
			os.Exit(1)
		}
	}
	if daemon {
		gdip.SetLogger(log.New(os.Stderr, "", log.LstdFlags))
	}

	g, err := gdip.New(serviceURL, username, password, domainName, mode, address)
	if err != nil {
		logFatal("Failed to initialize gdip client: %v", err)
	}

	if daemon {
		daemonLoop(g)
		return // Not reached
	}

	addr, err := g.Update()
	if err != nil {
		logError("Failed to updated: %v", err)
		os.Exit(1)
	}
	fmt.Println(addr)
	return
}
