package main

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"runtime"
)

const UNIVERSAL_PROXY = "localhost:5000"

// main file for load balancer
func main() {
	server_for_lb := []Server{ // Todo: this is to be replaced by GetServer and we take input from the user
		*SimpleServer("https://www.google.com"),
		*SimpleServer("https://www.bing.com"),
		*SimpleServer("https://www.youtube.com"),
		*SimpleServer("https://duckduckgo.com"),
	}

	if len(server_for_lb) <= 1 {
		fmt.Printf("No server working")
		return
	}
	lb := LoadBalancernew("8000", server_for_lb)
	handleredirect := func(w http.ResponseWriter, r *http.Request) {
		lb.ServeProxy(w, r)
	}
	http.HandleFunc("/", handleredirect)
	http.ListenAndServe(":"+lb.port, nil)
}

func getServer(w io.Writer) (*string, error) {
	input := ""

	fmt.Printf("Please paste or type your URL (press 'Esc' to exit):")
	fmt.Scan(input)
	return &input, nil
}

func clearTerminal(writer io.Writer) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	default:
		cmd = exec.Command("clear")
	}
	// Run the command and capture its output
	cmd.Stdout = writer
	cmd.Stderr = writer
	err := cmd.Run()
	if err != nil {
		writer.Write([]byte("Failed to clear terminal: " + err.Error()))
	}
}

func GetServer(writer io.Writer) []Server {
	var number_of_server int
	var server_to_return []Server
	fmt.Printf("Enter the number server to set up:")
	_, err := fmt.Scan(&number_of_server)
	if err != nil {
		var user int
		fmt.Printf("Error reading inputL %v\n Do you want to retry(1->yes or 0->no): ", err)
		_, err := fmt.Scan(&user)
		if err != nil && user != 1 {
			return []Server{}
		}
		getServer(writer)
	}
	for i := 0; i < number_of_server; i++ {
		url, err := getServer(writer)
		if err != nil {
			fmt.Printf("Error while getting server: %v\n", err)
		}
		server_to_return = append(server_to_return, *SimpleServer(*url))
	}
	return server_to_return
}
