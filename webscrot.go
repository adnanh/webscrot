package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"time"
)

var (
	workerPipe chan string
	mainPipe   chan bool
	jobDone    = false

	workerCount    = flag.Int("workers", 1, "number of concurrent workers")
	timeout        = flag.Int("timeout", 5000, "number of milliseconds to wait before taking a screenshot")
	width          = flag.Int("width", 1024, "screen width")
	height         = flag.Int("height", 768, "screen height")
	urlPrefix      = flag.String("urlprefix", "", "string to prefix the urls from JSON file")
	urlFile        = flag.String("urlfile", "url.json", "path to the JSON file containg an array of urls")
	path           = flag.String("outputpath", "./", "path where the screenshots should be saved")
	filenamePrefix = flag.String("fileprefix", "", "string to prefix the output filename")
	filenameSuffix = flag.String("filesuffix", "", "string to suffix the output filename")
)

type task []string

func prepareTasks() (*task, error) {
	tasks := &task{}
	file, e := ioutil.ReadFile(*urlFile)
	if e != nil {
		return nil, e
	}

	json.Unmarshal(file, tasks)
	return tasks, e
}

func processTask(id int) {
	idString := fmt.Sprintf(":%d", id+99)

	virtualDisplay := exec.Command("Xvfb", idString, "-screen", "0", fmt.Sprintf("%dx%dx16", *width, *height))
	ratpoison := exec.Command("ratpoison", "-d", idString)

	err := virtualDisplay.Start()
	if err != nil {
		fmt.Printf("[%d] Cannot start virtual display.\n%v+\n", id, err)
		virtualDisplay = nil
		goto DONE
	}

	time.Sleep(500 * time.Millisecond)

	err = ratpoison.Start()
	if err != nil {
		fmt.Printf("[%d] Cannot start ratpoison.\n%v+\n", id, err)
		ratpoison = nil
		goto CLEANUP
	}

	time.Sleep(500 * time.Millisecond)

	for {
		select {
		case url := <-workerPipe:
			fmt.Printf("[%d] Processing %s\n", id, url)
			browser := exec.Command("midori", "--display", idString, "-e", "Fullscreen", "-a", fmt.Sprintf("%s%s", *urlPrefix, url))
			err := browser.Start()
			if err != nil {
				fmt.Printf("[%d] Cannot start browser.\n%v+\n", id, err)
				mainPipe <- true
			} else {
				reg, _ := regexp.Compile("[^A-Za-z0-9]+")
				time.Sleep(time.Duration(*timeout) * time.Millisecond)
				screenshot := exec.Command("import", "-display", idString, "-window", "root", fmt.Sprintf("%s/%s%s%s.png", *path, *filenamePrefix, reg.ReplaceAllString(url, "-"), *filenameSuffix))
				screenshot.Dir, _ = os.Getwd()
				out, err := screenshot.CombinedOutput()
				if err != nil {
					fmt.Printf("[%d] Cannot take a screenshot.\n%v+\n%v+\n", id, string(out), err)
				}

				browser.Process.Kill()
				fmt.Printf("[%d] Finished processing %s\n", id, url)
				mainPipe <- true
			}
		default:
			if jobDone {
				goto CLEANUP
			}
		}
	}

CLEANUP:
	if virtualDisplay != nil {
		virtualDisplay.Process.Kill()
	}
	if ratpoison != nil {
		ratpoison.Process.Kill()
	}
DONE:
	mainPipe <- false
}

func main() {
	flag.Parse()

	workerPipe = make(chan string, *workerCount)

	tasks, err := prepareTasks()
	if err != nil {
		fmt.Printf("Error while trying to read %s: %v\n", *urlFile, err)
		os.Exit(1)
	}

	taskCount := len(*tasks)

	mainPipe = make(chan bool, taskCount+*workerCount)

	fmt.Printf("Starting up %d worker(s)\n", *workerCount)

	for i := 0; i < *workerCount; i++ {
		go processTask(i + 1)
	}

	fmt.Printf("Queueing up %d task(s)\n", taskCount)
	for i := 0; i < taskCount; i++ {
		workerPipe <- (*tasks)[i]
	}

	for {
		select {
		case _ = <-mainPipe:
			if jobDone {
				if taskCount--; taskCount == -*workerCount {
					fmt.Println("Bye!")
					os.Exit(0)
				}
			} else {
				if taskCount--; taskCount == 0 {
					fmt.Println("All done, terminating workers...")
					jobDone = true
				}
			}
		}
	}
}
