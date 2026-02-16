package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jamesphm04/WCP_detrack_monthly_report/internal/api"
)


func main() {
	fmt.Println("************RUNNING SANDBOX************")

	// Read jobs.json
	data, err := os.ReadFile("jobs.json")
	if err != nil {
		fmt.Println("Error reading jobs.json:", err)
		return
	}

	var jobs []api.Job
	if err := json.Unmarshal(data, &jobs); err != nil {
		fmt.Println("Failed to parse JSON:", err)
		return
	}

	fmt.Println("Total jobs:", len(jobs))

	for _, job := range(jobs) {
		if (job.RunNumber == "") {
			fmt.Println(job.DoNumber)
		}
	}
	
}
