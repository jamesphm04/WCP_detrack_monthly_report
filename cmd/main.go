package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jamesphm04/WCP_detrack_monthly_report/internal/api"
	"github.com/jamesphm04/WCP_detrack_monthly_report/internal/config"
	"github.com/jamesphm04/WCP_detrack_monthly_report/internal/logger"
	"github.com/jamesphm04/WCP_detrack_monthly_report/internal/notifier"
	"go.uber.org/zap"
)

type ReportEntry struct {
	RunNumber string `json:"run_number"`
	NumOrdersDelivered int `json:"num_orders_delivered"`
	NumPartsDelivered int `json:"num_parts_delivered"`
	NumOrdersPickedUp int `json:"num_orders_picked_up"`
	NumPartsPickedUp int `json:"num_parts_picked_up"`
	FreightRevenue float64 `json:"freight_revenue"`
}



func main() {
	// INIT
	// init logger
	log, err := logger.New()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	log.Info("Starting WCP Detrack Monthly Report app...")

	// init config
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// init Detrack client
	detrackClient := api.NewDetrackClient(log, cfg)

	// init Notifier
	notifier := notifier.NewNotifier(log, cfg)



	// MAIN
	jobs, err := detrackClient.GetJobs() // e.g., limit 1000
	if err != nil {
		log.Fatal("Failed to fetch jobs", zap.Error(err))
	}

	log.Info("Total jobs fetched", zap.Int("count", len(jobs)))

	// Save to json 
	jsonFile, err := os.Create("jobs.json")
	if err != nil {
		log.Fatal("Failed to create JSON file", zap.Error(err))
	}

	defer jsonFile.Close()

	// Encode jobs slice to JSON with indentation 
	encoder := json.NewEncoder(jsonFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(jobs); err != nil { //save and check for error
		log.Fatal("Failed to write JSON to file", zap.Error(err))
	}

	log.Info("Jobs saved to JSON file", zap.String("file", "jobs.json"))


	// Generate report.csv
	status := "completed"

	now := time.Now()
	year := now.Year()
	month := now.Month()
	loc, err := time.LoadLocation("Australia/Brisbane")
	if err != nil {
		log.Fatal("Failed to load Brisbane location", zap.Error(err))
	}
	fromDate := time.Date(year, month, 1, 0, 0, 0, 0, loc)
	toDate := fromDate.AddDate(0, 1, 0)

	log.Info(fmt.Sprintf("Processing jobs with Status: %s (%s - %s)\n", status, fromDate, toDate))

	// Aggregate report by run_number

	reportMap := make(map[string]*ReportEntry)

	for _, job := range jobs {
		// Filter by status
		if job.Status != status {
			continue
		}

		// filter by date
		formattedJobDate, err := time.ParseInLocation(
			"2006-01-02",
			job.Date,
			loc,
		)
		if err != nil ||
			formattedJobDate.Before(fromDate) ||
			!formattedJobDate.Before(toDate) {
			continue
		}

		// Append 
		entry, isExists := reportMap[job.RunNumber]
		freightFloat64, err := strconv.ParseFloat(job.JobPrice, 64)
		if err != nil {
			log.Error("Failed to parse Job Price",
				zap.String("jobID", job.ID),
				zap.Error(err),
			)
			freightFloat64 = 0
		}

		if !isExists {
			entry = &ReportEntry{
				RunNumber: job.RunNumber,
				FreightRevenue: freightFloat64, 
			}

			reportMap[job.RunNumber] = entry
		}

		if job.Type == "Delivery" {
			entry.NumOrdersDelivered++
			entry.NumPartsDelivered += int(job.ItemCount)
			entry.FreightRevenue += freightFloat64
		} else {
			entry.NumOrdersPickedUp++
			entry.NumPartsPickedUp += int(job.ItemCount)
			entry.FreightRevenue += freightFloat64
		}
	}


	// Convert map to slice an sort
	var reportSlice []ReportEntry
	for _, entry := range reportMap {
		reportSlice = append(reportSlice, *entry)
	}

	// Create a CSV File
	report_path := fmt.Sprintf("%d_%02d_detrack_monthly_report.csv", year, month)
	csvFile, err := os.Create(report_path)
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
			strconv.FormatFloat(r.FreightRevenue, 'f', 3, 64),
		})
	}

	// FLUSH AND CLOSE BEFORE SENDING EMAIL
	writer.Flush()
	if err := writer.Error(); err != nil {
		log.Fatal("Failed to flush CSV writer", zap.Error(err))
	}
	csvFile.Close()

	log.Info("CSV report generated successfully")

	// Send email 
	subject := "WCP Detrack Monthly Report Notification"
	body := fmt.Sprintf(
		"Hi,\n\nAttached is the report for Detrack from %s to %s.\n\nThanks",
		fromDate.Format("2006-01-02"),
		toDate.Format("2006-01-02"),
	)


	if err := notifier.Send(subject, body, []string{report_path}); err != nil {
		log.Error("Failed to send report email", zap.Error(err))
	} else {
		log.Info("Report email sent successfully")
	}

	log.Info("COMPLETED!")
}