// Author  Raido Pahtma
// License MIT

// Reportlogger executable.
package main

import "fmt"
import "os"
import "os/signal"
import "time"

import "github.com/jessevdk/go-flags"
import "github.com/proactivity-lab/go-loggers"
import "github.com/proactivity-lab/go-sfconnection"
import "github.com/thinnect/go-reportreceiver"

const ApplicationVersionMajor = 0
const ApplicationVersionMinor = 1
const ApplicationVersionPatch = 0

var ApplicationBuildDate string
var ApplicationBuildDistro string

type Options struct {
	Positional struct {
		ConnectionString string `description:"Connectionstring sf@HOST:PORT"`
	} `positional-args:"yes"`

	Output string `short:"o" long:"output" default:"reports.txt" description:"Reports output file"`

	Source sfconnection.AMAddr  `short:"s" long:"source" default:"0001" description:"Source of the packet (hex)"`
	Group  sfconnection.AMGroup `short:"g" long:"group" default:"22" description:"Packet AM Group (hex)"`

	Debug       []bool `short:"D" long:"debug"   description:"Debug mode, print raw packets"`
	ShowVersion func() `short:"V" long:"version" description:"Show application version"`
}

// Main function.
func mainfunction() int {

	var opts Options
	opts.ShowVersion = func() {
		if ApplicationBuildDate == "" {
			ApplicationBuildDate = "YYYY-mm-dd_HH:MM:SS"
		}
		if ApplicationBuildDistro == "" {
			ApplicationBuildDistro = "unknown"
		}
		fmt.Printf("reportlogger %d.%d.%d (%s %s)\n", ApplicationVersionMajor, ApplicationVersionMinor, ApplicationVersionPatch, ApplicationBuildDate, ApplicationBuildDistro)
		os.Exit(0)
	}

	_, err := flags.Parse(&opts)
	if err != nil {
		fmt.Printf("Argument parser error: %s\n", err)
		return 1
	}

	host, port, err := sfconnection.ParseSfConnectionString(opts.Positional.ConnectionString)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return 1
	}

	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, os.Kill)

	sfc := sfconnection.NewSfConnection()
	rl := reportreceiver.NewReportReceiver(sfc, opts.Source, opts.Group)
	rfw, _ := reportreceiver.NewReportFileWriter(opts.Output)
	rl.SetOutput(rfw)

	logger := loggers.BasicLogSetup(len(opts.Debug))
	if len(opts.Debug) > 0 {
		sfc.SetLoggers(logger)
	}
	rl.SetLoggers(logger)

	sfc.Autoconnect(host, port, 30*time.Second)

	go rl.Run()

	for interrupted := false; interrupted == false; {
		select {
		case sig := <-signals:
			signal.Stop(signals)
			logger.Debug.Printf("signal %s\n", sig)
			sfc.Disconnect()
			interrupted = true
			time.Sleep(time.Second)
		}
	}

	return 0
}

func main() {
	os.Exit(mainfunction())
}
