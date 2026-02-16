package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"

	"github.com/jamesphm04/WCP_detrack_monthly_report/internal/api"
	"github.com/jamesphm04/WCP_detrack_monthly_report/internal/config"
	"github.com/jamesphm04/WCP_detrack_monthly_report/internal/logger"
	"github.com/jamesphm04/WCP_detrack_monthly_report/internal/notifier"
	"github.com/jamesphm04/WCP_detrack_monthly_report/internal/processor"
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

	// init date range to report 

	// Load Brisbane timezone
	loc, err := time.LoadLocation("Australia/Brisbane")
	if err != nil {
		log.Fatal("Failed to load Brisbane location", zap.Error(err))
	}

	// Get today in Brisbane
	today := time.Now().In(loc)

	var fromDate, toDate time.Time
	mode := ""

	// -----------------------------
	// First Monday of the month → previous month
	// -----------------------------
	if today.Day() <= 7 {
		mode = "MONTH"
		firstOfThisMonth := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, loc)
		lastMonthEnd := firstOfThisMonth.AddDate(0, 0, -1) // last day previous month
		lastMonthStart := time.Date(lastMonthEnd.Year(), lastMonthEnd.Month(), 1, 0, 0, 0, 0, loc)

		fromDate = lastMonthStart
		toDate = lastMonthEnd
	} else {
		mode = "WEEK"
		// Previous week: Monday → Sunday
		// Find last week's Monday
		// today.Weekday() returns 0=Sunday, 1=Monday,...6=Saturday
		offset := int(today.Weekday()) // Mon=1, Tue=2...
		if offset == 0 {
			offset = 7 // Sunday → 7 days
		}

		// Last week's Monday = today - offset - 6
		lastMonday := today.AddDate(0, 0, -offset-6)
		lastSunday := lastMonday.AddDate(0, 0, 6)

		fromDate = time.Date(lastMonday.Year(), lastMonday.Month(), lastMonday.Day(), 0, 0, 0, 0, loc)
		toDate = time.Date(lastSunday.Year(), lastSunday.Month(), lastSunday.Day(), 23, 59, 59, 0, loc)
	}

	log.Info((fmt.Sprintf("MODE: %s, RANGE: %s - %s", mode, fromDate.Format("2006-01-02"), toDate.Format("2006-01-02"))))

	// init config
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// init Detrack client
	detrackClient := api.NewDetrackClient(log, cfg)

	// init Notifier
	notifier := notifier.NewNotifier(log, cfg)

	// init Excel file 
	f := excelize.NewFile()
    defer func() {
        if err := f.Close(); err != nil {
            fmt.Println(err)
        }
    }()
	
	// MAIN
	jobs, err := detrackClient.GetJobs() // e.g., limit 1000
	if err != nil {
		log.Fatal("Failed to fetch jobs", zap.Error(err))
	}

	// Preprocess jobs - normalize run numbers
	normalizer := processor.NewRunNumberNormalizer()
	for i := range jobs {
		jobs[i].RunNumber = normalizer.Normalize(jobs[i].RunNumber)
	}

	log.Info("Total jobs fetched", zap.Int("count", len(jobs)))

	// All Jobs sheet
	jobSheet := "Jobs"

	_, err = f.NewSheet(jobSheet)
	if err != nil {
		log.Fatal("Failed to create 'Jobs' sheet", zap.Error(err))
	}

	// Create Jobs headers
	jobHeaders := []string{
		"id",
		"status",
		"date",
		"type",
		"items_count",
		"job_price",
		"do_number",
		"run_number",
	}

	for i, h := range jobHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(jobSheet, cell, h)
	}

	// Pour Jobs data
	jobStartRow := 2
	for _, job := range jobs {
		f.SetCellValue(jobSheet, fmt.Sprintf("A%d", jobStartRow), job.ID)
		f.SetCellValue(jobSheet, fmt.Sprintf("B%d", jobStartRow), job.Status)
		f.SetCellValue(jobSheet, fmt.Sprintf("C%d", jobStartRow), job.Date)
		f.SetCellValue(jobSheet, fmt.Sprintf("D%d", jobStartRow), job.Type)
		f.SetCellValue(jobSheet, fmt.Sprintf("E%d", jobStartRow), job.ItemCount)
		f.SetCellValue(jobSheet, fmt.Sprintf("F%d", jobStartRow), job.JobPrice)
		f.SetCellValue(jobSheet, fmt.Sprintf("G%d", jobStartRow), job.DoNumber)
		f.SetCellValue(jobSheet, fmt.Sprintf("H%d", jobStartRow), job.RunNumber)
		jobStartRow++
	} 

	// Report Sheet
	status := "completed"

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
			log.Error("Failed to parse Job Price. Fallback to 0",
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

	// Report sheet
	reportSheet := "Report"
	_, err = f.NewSheet(reportSheet)
	if err != nil {
		log.Fatal("Failed to create 'Report' sheet", zap.Error(err))
	}

	// Delete default Sheet1 and set Report as active
	f.DeleteSheet("Sheet1")
	reportSheetIndex, err := f.GetSheetIndex(reportSheet)
	if err != nil {
		log.Error("Failed to get report sheet index", zap.Error(err))
	} else {
		f.SetActiveSheet(reportSheetIndex)
	}

	reportHeaders := []string {
		"run_number",
		"num_orders_delivered",
		"num_parts_delivered",
		"num_orders_picked_up",
		"num_parts_picked_up",
		"freight_revenue",
	}

	for i, h := range reportHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(reportSheet, cell, h)
	}

	// Pour Report data
	reportStartRow := 2
	for _, reportEntry := range reportSlice {
		f.SetCellValue(reportSheet, fmt.Sprintf("A%d", reportStartRow), reportEntry.RunNumber)
		f.SetCellValue(reportSheet, fmt.Sprintf("B%d", reportStartRow), reportEntry.NumOrdersDelivered)
		f.SetCellValue(reportSheet, fmt.Sprintf("C%d", reportStartRow), reportEntry.NumPartsDelivered)
		f.SetCellValue(reportSheet, fmt.Sprintf("D%d", reportStartRow), reportEntry.NumOrdersPickedUp)
		f.SetCellValue(reportSheet, fmt.Sprintf("E%d", reportStartRow), reportEntry.NumPartsPickedUp)
		f.SetCellValue(reportSheet, fmt.Sprintf("F%d", reportStartRow), reportEntry.FreightRevenue)
		reportStartRow++
	}

	// Save xlsx file 
	reportPath := fmt.Sprintf("detrack_report_%s_to_%s.xlsx",
		fromDate.Format("2006-01-02"),
		toDate.Format("2006-01-02"),
	)

	if err := f.SaveAs(reportPath); err != nil {
        log.Error("Failed to save XLSX file", zap.Error(err))
    }

	log.Info("XLSX report generated successfully")

	// Send email 
	subject := "WCP Detrack Monthly Report Notification"
	body := fmt.Sprintf(
		"Hi,\n\nAttached is the report for Detrack from %s to %s.\n\nThanks",
		fromDate.Format("2006-01-02"),
		toDate.Format("2006-01-02"),
	)


	if err := notifier.Send(subject, body, []string{reportPath}); err != nil {
		log.Error("Failed to send report email", zap.Error(err))
	} else {
		log.Info("Report email sent successfully")
	}

	log.Info("COMPLETED!")
}