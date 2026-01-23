package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/jamesphm04/WCP_detrack_monthly_report/internal/api"
)

type ReportEntry struct {
	RunNumber           string `json:"run_number"`
	NumOrdersDelivered  int    `json:"num_orders_delivered"`
	NumPartsDelivered   int    `json:"num_parts_delivered"`
	NumOrdersPickedUp   int    `json:"num_orders_picked_up"`
	NumPartsPickedUp    int    `json:"num_parts_picked_up"`
	FreightRevenue      int    `json:"freight_revenue"`
}

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

	STATUS := "completed"
	FROM_DATE := "2026-01-01"
	TO_DATE := "2026-02-01"

	layout := "2006-01-02"
	fromDate, _ := time.Parse(layout, FROM_DATE)
	toDate, _ := time.Parse(layout, TO_DATE)

	fmt.Printf("Processing jobs with Status: %s (%s - %s)\n", STATUS, FROM_DATE, TO_DATE)

	// Aggregate report by run_number
	reportMap := make(map[string]*ReportEntry)

	for _, job := range jobs {
		// Filter by status
		if job.Status != STATUS {
			continue
		}

		// Filter by date
		jobDate, err := time.Parse(layout, job.Date)
		if err != nil || jobDate.Before(fromDate) || jobDate.After(toDate) {
			continue
		}

		entry, exists := reportMap[job.RunNumber]
		if !exists {
			entry = &ReportEntry{
				RunNumber:      job.RunNumber,
				FreightRevenue: -1,
			}
			reportMap[job.RunNumber] = entry
		}

		parts := int(job.ItemCount)
		if parts == 0 {
			parts = 1
		}

		if job.Type == "Delivery" {
			entry.NumOrdersDelivered++
			entry.NumPartsDelivered += parts
		} else if job.Type == "Collection" {
			entry.NumOrdersPickedUp++
			entry.NumPartsPickedUp += parts
		}
	}

	// Convert map to slice and sort
	var reportSlice []ReportEntry
	for _, entry := range reportMap {
		reportSlice = append(reportSlice, *entry)
	}
	sort.Slice(reportSlice, func(i, j int) bool {
		return reportSlice[i].RunNumber < reportSlice[j].RunNumber
	})

	// Create CSV file
	csvFile, err := os.Create("report.csv")
	if err != nil {
		fmt.Println("Failed to create CSV file:", err)
		return
	}
	defer csvFile.Close()

	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	// Write header
	writer.Write([]string{
		"run_number",
		"num_orders_delivered",
		"num_parts_delivered",
		"num_orders_picked_up",
		"num_parts_picked_up",
		"freight_revenue",
	})

	// Write each report entry
	for _, r := range reportSlice {
		writer.Write([]string{
			r.RunNumber,
			strconv.Itoa(r.NumOrdersDelivered),
			strconv.Itoa(r.NumPartsDelivered),
			strconv.Itoa(r.NumOrdersPickedUp),
			strconv.Itoa(r.NumPartsPickedUp),
			strconv.Itoa(r.FreightRevenue),
		})
	}

	fmt.Println("CSV report generated successfully! Total runs:", len(reportSlice))
}
