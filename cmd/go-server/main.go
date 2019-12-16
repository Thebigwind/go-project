package main

import (
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
)

var (
	args_log_file = kingpin.Flag("log-file", "The log file. Default \"/var/log/project.log\"").
			Short('L').
			Default("/var/log/project.log").
			String()
	args_log_level = kingpin.Flag("log-level", "The log level. Default \"INFO\"").
			Short('l').
			Default("INFO").
			String()
	args_conf_file = kingpin.Flag("config-file", "The configure file. Default \"/etc/project.conf\"").
			Short('c').
			Default("/etc/project.conf").
			String()
)

func main() {
	fmt.Println("project server starts now")
	kingpin.CommandLine.HelpFlag.Short('h')
	kingpin.Parse()
	arguments := make(map[string]string)
	arguments["log-file"] = *args_log_file
	arguments["log-level"] = *args_log_level
	arguments["config-file"] = *args_conf_file
	err := ServerStart(arguments)
	if err != nil {
		fmt.Printf("Fail to start project server: %s\n",
			err.Error())
		os.Exit(1)
	}

}
