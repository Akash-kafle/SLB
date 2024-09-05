package main

import (
	"fmt"
	"net/url"

	"github.com/eiannone/keyboard"
)

const UNIVERSAL_PROXY = "localhost:5000"

// main file for load balancer

func main() {
	proxyurl, err := GetProxy()
	for err != nil {
		proxyurl, err = GetProxy()
	}
	if proxyurl == nil {
		for proxyurl == nil || err != nil {
			proxyurl, err = url.Parse(UNIVERSAL_PROXY)
		}
	}
}

func GetProxy() (*url.URL, error) {
	var input string

	// Initialize the keyboard listener
	if err := keyboard.Open(); err != nil {
		fmt.Println("Failed to open keyboard listener:", err)
		return nil, err
	}
	defer keyboard.Close()

	fmt.Println("Enter URL (Press Esc to cancel):")

	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			fmt.Println("Error reading key:", err)
			return nil, err
		}

		if key == keyboard.KeyEsc {
			fmt.Println("Esc key pressed. Exiting.")
			return nil, nil
		}

		if key == keyboard.KeyEnter {
			break
		}

		input += string(char)
		fmt.Print(string(char))
	}

	parsedURL, err := url.Parse(input)
	if err != nil {
		fmt.Println("Invalid URL. Please try again.")
		return GetProxy()
	}

	alive := IsBackendAlive(parsedURL)
	if !alive {
		fmt.Println("The provided URL's server is down. Please provide a new URL.")
		return GetProxy()
	}

	return parsedURL, nil
}
