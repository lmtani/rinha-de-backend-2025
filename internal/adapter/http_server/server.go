package http_server

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lmtani/rinha-de-backend-2025/internal/config"
	"github.com/lmtani/rinha-de-backend-2025/internal/domain"
	"github.com/lmtani/rinha-de-backend-2025/internal/usecase"
)

// Server handles HTTP requests for the payment application
type Server struct {
	requestPayment *usecase.RequestPaymentUseCase
	auditPayments  *usecase.AuditPaymentsUseCase
	engine         *gin.Engine
	config         *config.ServerConfig
}

// NewServer creates a new HTTP server instance
func NewServer(
	requestPayment *usecase.RequestPaymentUseCase,
	auditPayments *usecase.AuditPaymentsUseCase,
	cfg *config.ServerConfig,
) *Server {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())

	server := &Server{
		requestPayment: requestPayment,
		auditPayments:  auditPayments,
		engine:         engine,
		config:         cfg,
	}

	server.registerRoutes()
	return server
}

// registerRoutes sets up the HTTP routes
func (s *Server) registerRoutes() {
	s.engine.POST("/payments", s.handleRequestPayment)
	s.engine.GET("/payments-summary", s.handleAuditPayments)
	s.engine.GET("/health", s.handleHealth)
}

// Start starts the HTTP server
func (s *Server) Start() error {
	srv := &http.Server{
		Addr:              s.config.Port,
		Handler:           s.engine,
		ReadHeaderTimeout: s.config.ReadTimeout,
		WriteTimeout:      s.config.WriteTimeout,
	}

	log.Printf("HTTP server listening on %s", s.config.Port)
	return srv.ListenAndServe()
}

func (s *Server) handleRequestPayment(c *gin.Context) {
	var payment domain.Payment
	if err := c.ShouldBindJSON(&payment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
		return
	}

	if err := s.requestPayment.Execute(c.Request.Context(), payment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"status": "accepted"})
}

func (s *Server) handleAuditPayments(c *gin.Context) {
	// Optional query params: from, to (ISO 8601 in UTC)
	var fromPtr *time.Time
	var toPtr *time.Time

	if fromStr := c.Query("from"); fromStr != "" {
		// Accept RFC3339 format (e.g., 2020-07-10T12:34:56.000Z)
		t, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid 'from' timestamp, expected ISO8601 UTC (RFC3339)"})
			return
		}
		utc := t.UTC()
		fromPtr = &utc
	}

	if toStr := c.Query("to"); toStr != "" {
		t, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid 'to' timestamp, expected ISO8601 UTC (RFC3339)"})
			return
		}
		utc := t.UTC()
		toPtr = &utc
	}

	summary, err := s.auditPayments.Execute(c.Request.Context(), fromPtr, toPtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}
