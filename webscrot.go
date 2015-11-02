package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"time"
)

var (
	workerPipe chan string
	mainPipe   chan bool
	jobDone    = false

	workerCount         = flag.Int("workers", 1, "number of concurrent workers")
	delay               = flag.Int("delay", 5000, "number of milliseconds to wait before taking a screenshot")
	width               = flag.Int("width", 1024, "screen width")
	height              = flag.Int("height", 768, "screen height")
	urlPrefix           = flag.String("url-prefix", "", "string to prefix the input urls with")
	urlSuffix           = flag.String("url-suffix", "", "string to sufix the input urls with")
	urlFile             = flag.String("file", "-", "path to the input file")
	isJSON              = flag.Bool("json", false, "parse input file as JSON array of strings")
	outputPath          = flag.String("output-path", "./", "path where the screenshots should be saved")
	filenamePrefix      = flag.String("filename-prefix", "", "string to prefix the output filename with")
	filenameSuffix      = flag.String("filename-suffix", "", "string to suffix the output filename with")
	filenameExtension   = flag.String("filename-extension", "png", "filename extension to use for ImageMagick import command")
	displayNumberOffset = flag.Int("display-number-offset", 99, "number to offset display number for")
)

func prepareTasks() (*[]string, error) {
	var tasks []string

	var scanner *bufio.Scanner
	var err error

	if *urlFile == "-" {
		scanner = bufio.NewScanner(os.Stdin)
	} else {
		file, err := os.Open(*urlFile)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		scanner = bufio.NewScanner(file)
	}

	if *isJSON {
		var jsonInput []byte

		scanner.Split(bufio.ScanBytes)

		for scanner.Scan() {
			for _, b := range scanner.Bytes() {
				jsonInput = append(jsonInput, b)
			}
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}

		if err = json.Unmarshal(jsonInput, &tasks); err != nil {
			return nil, err
		}
	} else {
		for scanner.Scan() {
			tasks = append(tasks, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}

	return &tasks, err
}

func processTask(id int) {
	idString := fmt.Sprintf(":%d", id+*displayNumberOffset)

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
			browser := exec.Command("midori", "--display", idString, "-e", "Fullscreen", "-a", fmt.Sprintf("%s%s%s", *urlPrefix, url, *urlSuffix))
			err := browser.Start()
			if err != nil {
				fmt.Printf("[%d] Cannot start browser.\n%v+\n", id, err)
				mainPipe <- true
			} else {
				reg, _ := regexp.Compile("[^A-Za-z0-9]+")
				time.Sleep(time.Duration(*delay) * time.Millisecond)
				screenshot := exec.Command("import", "-display", idString, "-window", "root", fmt.Sprintf("%s/%s%s%s.%s", *outputPath, *filenamePrefix, reg.ReplaceAllString(url, "-"), *filenameSuffix, *filenameExtension))
				screenshot.Dir, _ = os.Getwd()
				out, err := screenshot.CombinedOutput()
				if err != nil {
					fmt.Printf("[%d] Cannot take a screenshot.\n%v+\n%v+\n", id, string(out), err)
				}

				browser.Process.Kill()
				browser.Wait()
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
	if ratpoison != nil {
		ratpoison.Process.Kill()
		ratpoison.Wait()
	}

	if virtualDisplay != nil {
		virtualDisplay.Process.Kill()
		virtualDisplay.Wait()
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
