package handler

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lmtani/rinha-de-backend-2025/internal/admin/client"
)

// AdminHandler handles admin web interface requests
type AdminHandler struct {
	defaultClient  *client.ProcessorClient
	fallbackClient *client.ProcessorClient
	templates      *template.Template
}

// ProcessorInfo holds information about a processor
type ProcessorInfo struct {
	Name        string
	URL         string
	Client      *client.ProcessorClient
	Status      string
	LastChecked time.Time
}

// DashboardData holds data for the dashboard template
type DashboardData struct {
	Processors []ProcessorInfo
	Token      string
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(defaultURL, fallbackURL, token string) *AdminHandler {
	// Load templates
	templates := template.Must(template.ParseGlob("web/templates/*.html"))

	return &AdminHandler{
		defaultClient:  client.NewProcessorClient(defaultURL, token),
		fallbackClient: client.NewProcessorClient(fallbackURL, token),
		templates:      templates,
	}
}

// RegisterRoutes registers all admin routes
func (h *AdminHandler) RegisterRoutes(r *gin.Engine) {
	// Serve static files
	r.Static("/static", "web/static")

	// Main dashboard
	r.GET("/", h.dashboard)

	// Processor management
	r.GET("/processor/:name/summary", h.getProcessorSummary)
	r.POST("/processor/:name/token", h.setProcessorToken)
	r.POST("/processor/:name/delay", h.setProcessorDelay)
	r.POST("/processor/:name/failure", h.setProcessorFailure)
	r.POST("/processor/:name/purge", h.purgeProcessorPayments)

	// Global actions
	r.POST("/global/token", h.setGlobalToken)
	r.POST("/global/delay", h.setGlobalDelay)
	r.POST("/global/failure", h.setGlobalFailure)
	r.POST("/global/purge", h.purgeAllPayments)
}

// dashboard renders the main dashboard
func (h *AdminHandler) dashboard(c *gin.Context) {
	processors := []ProcessorInfo{
		{
			Name:   "default",
			URL:    h.defaultClient.BaseURL,
			Client: h.defaultClient,
		},
		{
			Name:   "fallback",
			URL:    h.fallbackClient.BaseURL,
			Client: h.fallbackClient,
		},
	}

	// Check processor status
	ctx := context.Background()
	for i := range processors {
		_, err := processors[i].Client.GetPaymentsSummary(ctx, nil, nil)
		if err != nil {
			processors[i].Status = "Error: " + err.Error()
		} else {
			processors[i].Status = "OK"
		}
		processors[i].LastChecked = time.Now()
	}

	data := DashboardData{
		Processors: processors,
		Token:      h.defaultClient.Token,
	}

	if err := h.templates.ExecuteTemplate(c.Writer, "dashboard.html", data); err != nil {
		c.String(http.StatusInternalServerError, "Template error: %v", err)
		return
	}
}

// getProcessorSummary returns the payments summary for a specific processor
func (h *AdminHandler) getProcessorSummary(c *gin.Context) {
	processorName := c.Param("name")
	client := h.getClient(processorName)
	if client == nil {
		c.String(http.StatusBadRequest, "Invalid processor name")
		return
	}

	// Parse optional query parameters
	var from, to *time.Time
	if fromStr := c.Query("from"); fromStr != "" {
		if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
			from = &t
		}
	}
	if toStr := c.Query("to"); toStr != "" {
		if t, err := time.Parse(time.RFC3339, toStr); err == nil {
			to = &t
		}
	}

	ctx := context.Background()
	summary, err := client.GetPaymentsSummary(ctx, from, to)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error: %v", err)
		return
	}

	// Render summary partial template
	if err := h.templates.ExecuteTemplate(c.Writer, "summary.html", map[string]interface{}{
		"Processor": processorName,
		"Summary":   summary,
		"Now":       time.Now().Format("2006-01-02 15:04:05"),
	}); err != nil {
		c.String(http.StatusInternalServerError, "Template error: %v", err)
	}
}

// setProcessorToken sets token for a specific processor
func (h *AdminHandler) setProcessorToken(c *gin.Context) {
	processorName := c.Param("name")
	token := c.PostForm("token")

	if token == "" {
		c.String(http.StatusBadRequest, "Token is required")
		return
	}

	client := h.getClient(processorName)
	if client == nil {
		c.String(http.StatusBadRequest, "Invalid processor name")
		return
	}

	ctx := context.Background()
	if err := client.SetToken(ctx, token); err != nil {
		c.String(http.StatusInternalServerError, "Error: %v", err)
		return
	}

	// Update client token
	client.UpdateToken(token)

	c.String(http.StatusOK, "Token updated successfully for %s", processorName)
}

// setProcessorDelay sets delay for a specific processor
func (h *AdminHandler) setProcessorDelay(c *gin.Context) {
	processorName := c.Param("name")
	delayStr := c.PostForm("delay")

	delay, err := strconv.Atoi(delayStr)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid delay value")
		return
	}

	client := h.getClient(processorName)
	if client == nil {
		c.String(http.StatusBadRequest, "Invalid processor name")
		return
	}

	ctx := context.Background()
	if err := client.SetDelay(ctx, delay); err != nil {
		c.String(http.StatusInternalServerError, "Error: %v", err)
		return
	}

	c.String(http.StatusOK, "Delay set to %d ms for %s", delay, processorName)
}

// setProcessorFailure sets failure mode for a specific processor
func (h *AdminHandler) setProcessorFailure(c *gin.Context) {
	processorName := c.Param("name")

	// Try to get from form data first, then from JSON body
	failureStr := c.PostForm("failure")
	if failureStr == "" {
		// If not in form data, try to parse from JSON body
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err == nil {
			if val, exists := data["failure"]; exists {
				switch v := val.(type) {
				case string:
					failureStr = v
				case bool:
					if v {
						failureStr = "true"
					} else {
						failureStr = "false"
					}
				}
			}
		}
	}

	// Log the received data for debugging
	fmt.Printf("Processor: %s, Failure string: '%s'\n", processorName, failureStr)

	failure := failureStr == "true" || failureStr == "1"

	client := h.getClient(processorName)
	if client == nil {
		c.String(http.StatusBadRequest, "Invalid processor name")
		return
	}

	ctx := context.Background()
	if err := client.SetFailure(ctx, failure); err != nil {
		fmt.Printf("Error setting failure mode: %v\n", err)
		c.String(http.StatusInternalServerError, "Error: %v", err)
		return
	}

	status := "disabled"
	if failure {
		status = "enabled"
	}
	c.String(http.StatusOK, "Failure mode %s for %s", status, processorName)
}

// purgeProcessorPayments purges payments for a specific processor
func (h *AdminHandler) purgeProcessorPayments(c *gin.Context) {
	processorName := c.Param("name")

	client := h.getClient(processorName)
	if client == nil {
		c.String(http.StatusBadRequest, "Invalid processor name")
		return
	}

	ctx := context.Background()
	response, err := client.PurgePayments(ctx)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error: %v", err)
		return
	}

	c.String(http.StatusOK, "%s: %s", processorName, response.Message)
}

// Global actions that apply to both processors

func (h *AdminHandler) setGlobalToken(c *gin.Context) {
	token := c.PostForm("token")

	if token == "" {
		c.String(http.StatusBadRequest, "Token is required")
		return
	}

	ctx := context.Background()
	var errors []string

	// Update default processor
	if err := h.defaultClient.SetToken(ctx, token); err != nil {
		errors = append(errors, fmt.Sprintf("Default: %v", err))
	} else {
		h.defaultClient.UpdateToken(token)
	}

	// Update fallback processor
	if err := h.fallbackClient.SetToken(ctx, token); err != nil {
		errors = append(errors, fmt.Sprintf("Fallback: %v", err))
	} else {
		h.fallbackClient.UpdateToken(token)
	}

	if len(errors) > 0 {
		c.String(http.StatusInternalServerError, "Errors: %v", errors)
		return
	}

	c.String(http.StatusOK, "Token updated successfully for all processors")
}

func (h *AdminHandler) setGlobalDelay(c *gin.Context) {
	delayStr := c.PostForm("delay")

	delay, err := strconv.Atoi(delayStr)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid delay value")
		return
	}

	ctx := context.Background()
	var errors []string

	if err := h.defaultClient.SetDelay(ctx, delay); err != nil {
		errors = append(errors, fmt.Sprintf("Default: %v", err))
	}

	if err := h.fallbackClient.SetDelay(ctx, delay); err != nil {
		errors = append(errors, fmt.Sprintf("Fallback: %v", err))
	}

	if len(errors) > 0 {
		c.String(http.StatusInternalServerError, "Errors: %v", errors)
		return
	}

	c.String(http.StatusOK, "Delay set to %d ms for all processors", delay)
}

func (h *AdminHandler) setGlobalFailure(c *gin.Context) {
	// Try to get from form data first, then from JSON body
	failureStr := c.PostForm("failure")
	if failureStr == "" {
		// If not in form data, try to parse from JSON body
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err == nil {
			if val, exists := data["failure"]; exists {
				switch v := val.(type) {
				case string:
					failureStr = v
				case bool:
					if v {
						failureStr = "true"
					} else {
						failureStr = "false"
					}
				}
			}
		}
	}

	failure := failureStr == "true" || failureStr == "1"

	ctx := context.Background()
	var errors []string

	if err := h.defaultClient.SetFailure(ctx, failure); err != nil {
		errors = append(errors, fmt.Sprintf("Default: %v", err))
	}

	if err := h.fallbackClient.SetFailure(ctx, failure); err != nil {
		errors = append(errors, fmt.Sprintf("Fallback: %v", err))
	}

	if len(errors) > 0 {
		c.String(http.StatusInternalServerError, "Errors: %v", errors)
		return
	}

	status := "disabled"
	if failure {
		status = "enabled"
	}
	c.String(http.StatusOK, "Failure mode %s for all processors", status)
}

func (h *AdminHandler) purgeAllPayments(c *gin.Context) {
	ctx := context.Background()
	var results []string

	// Purge default processor
	if response, err := h.defaultClient.PurgePayments(ctx); err != nil {
		results = append(results, fmt.Sprintf("Default: Error - %v", err))
	} else {
		results = append(results, fmt.Sprintf("Default: %s", response.Message))
	}

	// Purge fallback processor
	if response, err := h.fallbackClient.PurgePayments(ctx); err != nil {
		results = append(results, fmt.Sprintf("Fallback: Error - %v", err))
	} else {
		results = append(results, fmt.Sprintf("Fallback: %s", response.Message))
	}

	c.String(http.StatusOK, "Results: %v", results)
}

// getClient returns the appropriate client based on processor name
func (h *AdminHandler) getClient(name string) *client.ProcessorClient {
	switch name {
	case "default":
		return h.defaultClient
	case "fallback":
		return h.fallbackClient
	default:
		return nil
	}
}
