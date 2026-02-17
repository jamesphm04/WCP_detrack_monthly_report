package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jamesphm04/WCP_detrack_monthly_report/internal/config"
	"go.uber.org/zap"
)

type Job struct {
	ID 			string `json:"id"`
	Status 		string `json:"status"` // 	Job Status. completed
	Date 		string `json:"date"` // Date for performing the job. 2019-12-24
	Type        string `json:"type"` // Detrack Job Type Delivery/Collection. Delivery
	ItemCount 		float32 `json:"items_count"` // Number of entries in the Item Details list. 10
	JobPrice 		string `json:"job_price"` // Price of the job. "10.34"
	// TotalPrice 		string `json:"total_price"` // Total price amount for the job. 100
	// InvoiceAmount 		float32 `json:"invoice_amount"` // The amount for the job invoice. 1.5
	// PaymentAmount 		float32 `json:"payment_amount"` // The amount to be collected for the job. 1.5
	DoNumber 	string `json:"do_number"` // Unique identifier for the job. DO123
	// InvoiceNumber 		string `json:"invoice_number"` // The invoice number of the job. Inv123
	RunNumber   string `json:"run_number"` // The run number which the job belongs to. 1
}

// DetrackClient handle API requests
type DetrackClient struct {
	BaseURL 	string
	APIKey 		string
	FetchLimit  int
	HTTPClient 	*http.Client
	Logger 		*zap.Logger
}

// NewDetrackClient constructor
func NewDetrackClient(
	logger *zap.Logger,
	cfg *config.Config,
) *DetrackClient {
	return &DetrackClient{
		BaseURL: cfg.BaseURL,
		APIKey: cfg.APIKey,
		FetchLimit: cfg.FetchLimit,
		HTTPClient: &http.Client{Timeout: 60 * time.Second},
		Logger: logger,
	}
}

// GetJobs fetches all of the existing jobs on Detrack
func (c *DetrackClient) GetJobs() ([]Job, error) {
	c.Logger.Info("Getting all of the existing Jobs on Detrack...")

	allJobs := []Job{}
	url := fmt.Sprintf(
		"%s/dn/jobs?limit=%d",
		c.BaseURL,
		c.FetchLimit,
	)

	for url != "" {
		// Build a request
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			c.Logger.Error("HTTP request failed", zap.Error(err))
			return nil, err
		}

		req.Header.Set("X-API-KEY", c.APIKey)
		req.Header.Set("Content-Type", "application/json")

		// Execute request
		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			c.Logger.Error("Failed to read response body", zap.Error(err))
			return nil, err
		}

		// Read response
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			c.Logger.Error("Failed to read response body", zap.Error(err))
			return nil, err
		}

		// // Log the raw JSON response
		// rawJSON := string(body)
		// c.Logger.Debug("Raw API response", zap.String("raw", rawJSON))

		if resp.StatusCode != http.StatusOK {
			c.Logger.Error("API returned non-200 status",
				zap.Int("status", resp.StatusCode),
				zap.String("statusText", resp.Status))
			break
		}

		// Response structure
		var result struct {
			Data []Job 					`json:"data"`
			Links map[string]string 	`json:"links"`
		}

		if err := json.Unmarshal(body, &result); err != nil {
			c.Logger.Error("Failed to unmarshal JSON", zap.Error(err))
		}
		
		allJobs = append(allJobs, result.Data...)
		c.Logger.Info("Retrieved jobs so far", zap.Int("count", len(allJobs)))

		// Pagination 
		nextLink, ok := result.Links["next"]
		if ok && nextLink != "" {
			if nextLink[:4] != "http" {
				nextLink = c.BaseURL + nextLink
			}
			url = nextLink
		} else {
			url = ""
		}

	}

	c.Logger.Info("Finished fetching all jobs", zap.Int("total", len(allJobs)))
	return allJobs, nil
}