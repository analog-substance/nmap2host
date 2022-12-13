package nmap

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/analog-substance/arsenic/lib/host"
)

func NmapSplit(path string, name string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	outputName := fmt.Sprintf("%s.nmap", name)

	doneRe := regexp.MustCompile(`# Nmap done`)
	hostRe := regexp.MustCompile(`Nmap scan report for ([^ ]+)(?: \(([0-9\.]+)\))?$`)

	var currentHost *host.Host
	var lines []string
	commandLine := ""
	hostname := ""
	ip := ""

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if commandLine == "" {
			commandLine = line
			continue
		}

		// If we hit this string, there are no more hosts
		if doneRe.MatchString(line) {
			err = writeToFile(currentHost, outputName, []byte(strings.Join(lines, "\n")))
			if err != nil {
				return err
			}
			break
		}

		if hostRe.MatchString(line) {
			// Write nmap file for current host before starting a new host
			if currentHost != nil {
				err = writeToFile(currentHost, outputName, []byte(strings.Join(lines, "\n")))
				if err != nil {
					return err
				}
			}

			ip = ""
			hostname = ""

			match := hostRe.FindStringSubmatch(line)
			if len(match) == 2 {
				ip = match[1]
			} else if len(match) == 3 {
				hostname = match[1]
				ip = match[2]
			}

			currentHost, err = getHost([]string{hostname}, []string{ip})
			if err != nil {
				return err
			}

			lines = []string{commandLine}
			continue
		}

		lines = append(lines, line)
	}

	return nil
}
