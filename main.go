package main

import (
	"fmt"
	"github.com/playwright-community/playwright-go"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
)

func worker(id int, wg *sync.WaitGroup, stopCh <-chan struct{}, tabCount int) {
	defer wg.Done()

	log.Printf("worker %d", id)
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}
	browser, err := pw.Chromium.Launch()
	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}
	p, err := browser.NewContext(playwright.BrowserNewContextOptions{JavaScriptEnabled: playwright.Bool(true)})
	if err != nil {
		log.Fatalf("could not create page: %v", err)
	}

	for i := 0; i < tabCount; i++ {
		_, err = p.NewPage()
		if err != nil {
			log.Fatalf("could not create page: %v", err)
		}
	}

	//_, err = p.NewPage()
	//if err != nil {
	//	log.Fatalf("could not create page: %v", err)
	//}

	<-stopCh
	//fmt.Printf("Worker %d shutting down\n", id)
	//if err = browser.Close(); err != nil {
	//	log.Fatalf("could not close browser: %v", err)
	//}
	//if err = pw.Stop(); err != nil {
	//	log.Fatalf("could not stop Playwright: %v", err)
	//}
	return

}

func main() {

	args := os.Args
	if len(args) < 3 {
		log.Fatalf("no cli arguments")
	}
	numOfBrowsers, _ := strconv.Atoi(args[1])
	numOfTabs, _ := strconv.Atoi(args[2])

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	// Create a channel to signal the workers to stop
	stopChs := make([]chan struct{}, numOfBrowsers)

	// Use a WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Start multiple worker goroutines
	for i := 0; i < numOfBrowsers; i++ {
		stopChs[i] = make(chan struct{})
		wg.Add(1)
		go worker(i, &wg, stopChs[i], numOfTabs)
	}

	// Listen for the signal
	go func() {
		<-signalCh
		fmt.Println("\nReceived SIGINT. Initiating graceful shutdown...")
		for i := 0; i < numOfBrowsers; i++ {
			close(stopChs[i])
		}
	}()

	// Wait for all workers to finish
	wg.Wait()

	fmt.Println("All workers have completed. Exiting.")

}
