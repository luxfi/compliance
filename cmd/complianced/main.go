// Copyright 2024-2026 Lux Partners Limited. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

// Command complianced is a standalone HTTP server exposing the compliance module
// as a REST API. It uses only the standard library for HTTP routing.
package main

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/luxfi/compliance/pkg/aml"
	"github.com/luxfi/compliance/pkg/idv"
	"github.com/luxfi/compliance/pkg/kyc"
	"github.com/luxfi/compliance/pkg/onboarding"
	"github.com/luxfi/compliance/pkg/payments"
	"github.com/luxfi/compliance/pkg/rbac"
	"github.com/luxfi/compliance/pkg/regulatory"
	"github.com/luxfi/compliance/pkg/store"
	"github.com/luxfi/compliance/pkg/types"
	"github.com/luxfi/compliance/pkg/webhook"
)

func main() {
	logger := log.New(os.Stderr, "", log.LstdFlags)

	// --- Configuration from environment ---
	listen := envOr("COMPLIANCE_LISTEN", ":8091")
	apiKey := os.Getenv("COMPLIANCE_API_KEY")
	if apiKey == "" {
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			logger.Fatalf("FATAL: failed to generate random API key: %v", err)
		}
		apiKey = hex.EncodeToString(b)
		logger.Printf("WARNING: COMPLIANCE_API_KEY not set — generated random key: %s", apiKey)
	}
	defaultProvider := envOr("KYC_DEFAULT_PROVIDER", "jumio")

	// --- Initialize services ---
	kycService := kyc.NewService()
	appStore := kyc.NewStore()
	screeningService := aml.NewScreeningService(aml.DefaultScreeningConfig())
	monitoringService := aml.NewMonitoringService()
	paymentEngine := payments.NewComplianceEngine(screeningService)
	webhookHandler := webhook.NewHandler()

	// Initialize compliance store (in-memory; PostgresStore via DATABASE_URL in production).
	complianceStore := store.NewMemoryStore()

	// Seed default RBAC roles.
	for _, role := range rbac.DefaultRoles() {
		complianceStore.SaveRole(role)
	}

	// Ensure onboarding package is wired (used by /v2/applications endpoints).
	_ = onboarding.NewApplicationSteps
	_ = types.AppDraft

	// Register IDV providers from environment
	registerProviders(kycService, webhookHandler, defaultProvider)

	// Install default monitoring rules
	installDefaultRules(monitoringService)

	// --- Build router ---
	mux := http.NewServeMux()

	// Health check (no auth)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// KYC: initiate verification
	mux.HandleFunc("/v1/kyc/verify", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		var req idv.VerificationRequest
		if err := decodeBody(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		resp, err := kycService.InitiateKYC(r.Context(), &req)
		if err != nil {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, resp)
	})

	// KYC: webhook (no API key auth — providers can't send it)
	mux.HandleFunc("/v1/kyc/webhook/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		provider := lastSegment(r.URL.Path)
		if provider == "" {
			writeError(w, http.StatusBadRequest, "missing provider in path")
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeError(w, http.StatusBadRequest, "failed to read body")
			return
		}
		headers := flattenHeaders(r.Header)

		// Try the unified webhook handler first
		result, whErr := webhookHandler.Handle(provider, body, headers)
		if whErr != nil && result == nil {
			writeError(w, http.StatusBadRequest, whErr.Error())
			return
		}

		// Also pass through to KYC service for state updates
		event, kycErr := kycService.HandleWebhook(provider, body, headers)
		if kycErr != nil {
			// If webhook handler accepted it, still return success
			if result != nil {
				writeJSON(w, http.StatusOK, result)
				return
			}
			writeError(w, http.StatusBadRequest, kycErr.Error())
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"webhook": result,
			"event":   event,
		})
	})

	// KYC: get verification status
	mux.HandleFunc("/v1/kyc/status/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		verificationID := lastSegment(r.URL.Path)
		if verificationID == "" {
			writeError(w, http.StatusBadRequest, "missing verification ID")
			return
		}
		v, err := kycService.GetStatus(verificationID)
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, v)
	})

	// KYC: get verifications for application
	mux.HandleFunc("/v1/kyc/application/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		appID := lastSegment(r.URL.Path)
		if appID == "" {
			writeError(w, http.StatusBadRequest, "missing application ID")
			return
		}
		verifications := kycService.GetByApplication(appID)
		writeJSON(w, http.StatusOK, verifications)
	})

	// Applications: stats must be registered BEFORE the /{id} catch-all
	mux.HandleFunc("/v1/applications/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		writeJSON(w, http.StatusOK, appStore.Stats())
	})

	// Applications: create + list
	mux.HandleFunc("/v1/applications", func(w http.ResponseWriter, r *http.Request) {
		// Exact match only — requests with trailing path segments go to the next handler
		if r.URL.Path != "/v1/applications" {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		switch r.Method {
		case http.MethodPost:
			var app kyc.Application
			if err := decodeBody(r, &app); err != nil {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			if err := appStore.Create(&app); err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			writeJSON(w, http.StatusCreated, app)
		case http.MethodGet:
			status := r.URL.Query().Get("status")
			writeJSON(w, http.StatusOK, appStore.List(status))
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})

	// Applications: get, update, submit by ID
	mux.HandleFunc("/v1/applications/", func(w http.ResponseWriter, r *http.Request) {
		// Strip prefix to get remaining path
		rest := strings.TrimPrefix(r.URL.Path, "/v1/applications/")
		if rest == "" || rest == "stats" {
			// Already handled above; this shouldn't normally fire due to registration
			// order, but guard anyway.
			writeError(w, http.StatusNotFound, "not found")
			return
		}

		// Check for /v1/applications/{id}/submit
		parts := strings.SplitN(rest, "/", 2)
		appID := parts[0]

		if len(parts) == 2 && parts[1] == "submit" {
			if r.Method != http.MethodPost {
				writeError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
			app, err := appStore.Get(appID)
			if err != nil {
				writeError(w, http.StatusNotFound, err.Error())
				return
			}
			if app.Status != kyc.StatusDraft {
				writeError(w, http.StatusConflict, fmt.Sprintf("application is %s, not draft", app.Status))
				return
			}
			now := time.Now()
			app.Status = kyc.StatusPending
			app.SubmittedAt = &now
			if err := appStore.Update(app); err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, app)
			return
		}

		// Only bare ID — no extra path segments
		if len(parts) > 1 {
			writeError(w, http.StatusNotFound, "not found")
			return
		}

		switch r.Method {
		case http.MethodGet:
			app, err := appStore.Get(appID)
			if err != nil {
				writeError(w, http.StatusNotFound, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, app)
		case http.MethodPatch:
			existing, err := appStore.Get(appID)
			if err != nil {
				writeError(w, http.StatusNotFound, err.Error())
				return
			}
			// Decode partial update on top of existing
			if err := decodeBody(r, existing); err != nil {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			existing.ID = appID // prevent ID override
			if err := appStore.Update(existing); err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, existing)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})

	// AML: screen individual
	mux.HandleFunc("/v1/aml/screen", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		var req aml.ScreeningRequest
		if err := decodeBody(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		result, err := screeningService.Screen(r.Context(), &req)
		if err != nil {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, result)
	})

	// AML: monitor transaction
	mux.HandleFunc("/v1/aml/monitor", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		var tx aml.Transaction
		if err := decodeBody(r, &tx); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		result, err := monitoringService.Monitor(r.Context(), &tx)
		if err != nil {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, result)
	})

	// AML: list alerts
	mux.HandleFunc("/v1/aml/alerts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		status := r.URL.Query().Get("status")
		writeJSON(w, http.StatusOK, monitoringService.GetAlerts(status))
	})

	// Payments: validate
	mux.HandleFunc("/v1/payments/validate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		var req payments.PaymentRequest
		if err := decodeBody(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		var result *payments.PaymentResult
		var err error
		switch req.Direction {
		case payments.PaymentPayout:
			result, err = paymentEngine.ValidatePayout(r.Context(), &req)
		default:
			result, err = paymentEngine.ValidatePayin(r.Context(), &req)
		}
		if err != nil {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, result)
	})

	// Regulatory: get jurisdiction requirements
	mux.HandleFunc("/v1/regulatory/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		code := lastSegment(r.URL.Path)
		if code == "" {
			writeError(w, http.StatusBadRequest, "missing jurisdiction code")
			return
		}
		j := regulatory.GetJurisdiction(strings.ToUpper(code))
		if j == nil {
			writeError(w, http.StatusNotFound, fmt.Sprintf("jurisdiction %q not supported", code))
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"name":         j.Name(),
			"code":         j.Code(),
			"requirements": j.Requirements(),
			"limits":       j.TransactionLimits(),
		})
	})

	// Providers: list registered IDV providers
	mux.HandleFunc("/v1/providers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"providers": kycService.ListProviders(),
		})
	})

	// --- Compliance Store Endpoints (v2) ---

	// Dashboard
	mux.HandleFunc("/v2/dashboard", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		writeJSON(w, http.StatusOK, complianceStore.ComputeDashboard())
	})

	// Roles
	mux.HandleFunc("/v2/roles", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		writeJSON(w, http.StatusOK, complianceStore.ListRoles())
	})

	// Modules (permission matrix)
	mux.HandleFunc("/v2/modules", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		writeJSON(w, http.StatusOK, rbac.ComplianceModules())
	})

	// --- Middleware stack ---
	var handler http.Handler = mux
	handler = loggingMiddleware(logger, handler)
	handler = apiKeyMiddleware(apiKey, handler)

	// --- Server ---
	srv := &http.Server{
		Addr:         listen,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Printf("complianced listening on %s", listen)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("listen error: %v", err)
		}
	}()

	<-ctx.Done()
	logger.Println("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatalf("shutdown error: %v", err)
	}
	logger.Println("stopped")
}

// --- Middleware ---

// apiKeyMiddleware validates the X-Api-Key header. Skips /healthz and webhook paths.
func apiKeyMiddleware(key string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health check and webhook endpoints
		if r.URL.Path == "/healthz" || strings.HasPrefix(r.URL.Path, "/v1/kyc/webhook/") {
			next.ServeHTTP(w, r)
			return
		}
		// Require API key — no empty-key bypass
		if key == "" {
			writeError(w, http.StatusServiceUnavailable, "COMPLIANCE_API_KEY not configured")
			return
		}
		provided := r.Header.Get("X-Api-Key")
		if subtle.ConstantTimeCompare([]byte(provided), []byte(key)) != 1 {
			writeError(w, http.StatusUnauthorized, "invalid or missing API key")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs method, path, status, and duration to stderr.
func loggingMiddleware(logger *log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		logger.Printf("%s %s %d %s", r.Method, r.URL.Path, rw.status, time.Since(start).Round(time.Microsecond))
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
	wrote  bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.wrote {
		rw.status = code
		rw.wrote = true
	}
	rw.ResponseWriter.WriteHeader(code)
}

// --- Helpers ---

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func decodeBody(r *http.Request, v interface{}) error {
	if r.Body == nil {
		return fmt.Errorf("request body is empty")
	}
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	return nil
}

// lastSegment returns the final path segment after the last "/".
func lastSegment(path string) string {
	path = strings.TrimSuffix(path, "/")
	if i := strings.LastIndex(path, "/"); i >= 0 {
		return path[i+1:]
	}
	return path
}

// flattenHeaders converts http.Header to a simple map (first value only).
func flattenHeaders(h http.Header) map[string]string {
	m := make(map[string]string, len(h))
	for k, v := range h {
		if len(v) > 0 {
			m[k] = v[0]
		}
	}
	return m
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// --- Provider registration ---

func registerProviders(kycSvc *kyc.Service, whHandler *webhook.Handler, defaultProvider string) {
	// Jumio
	if token := os.Getenv("JUMIO_API_TOKEN"); token != "" {
		p, err := idv.GetProvider(idv.ProviderJumio, map[string]string{
			"api_token":  token,
			"api_secret": os.Getenv("JUMIO_API_SECRET"),
		})
		if err == nil {
			kycSvc.RegisterProvider(p)
			if secret := os.Getenv("JUMIO_WEBHOOK_SECRET"); secret != "" {
				kycSvc.SetWebhookSecret(idv.ProviderJumio, secret)
				whHandler.RegisterProvider(webhook.WebhookConfig{
					Provider:        idv.ProviderJumio,
					Secret:          secret,
					SignatureHeader: "X-Jumio-Signature",
				}, makeKYCProcessor(kycSvc, idv.ProviderJumio))
			}
		}
	}

	// Onfido
	if token := os.Getenv("ONFIDO_API_TOKEN"); token != "" {
		p, err := idv.GetProvider(idv.ProviderOnfido, map[string]string{
			"api_token":     token,
			"webhook_token": os.Getenv("ONFIDO_WEBHOOK_SECRET"),
		})
		if err == nil {
			kycSvc.RegisterProvider(p)
			if secret := os.Getenv("ONFIDO_WEBHOOK_SECRET"); secret != "" {
				kycSvc.SetWebhookSecret(idv.ProviderOnfido, secret)
				whHandler.RegisterProvider(webhook.WebhookConfig{
					Provider:        idv.ProviderOnfido,
					Secret:          secret,
					SignatureHeader: "X-SHA2-Signature",
				}, makeKYCProcessor(kycSvc, idv.ProviderOnfido))
			}
		}
	}

	// Plaid
	if clientID := os.Getenv("PLAID_CLIENT_ID"); clientID != "" {
		p, err := idv.GetProvider(idv.ProviderPlaid, map[string]string{
			"client_id": clientID,
			"secret":    os.Getenv("PLAID_SECRET"),
		})
		if err == nil {
			kycSvc.RegisterProvider(p)
			if secret := os.Getenv("PLAID_WEBHOOK_SECRET"); secret != "" {
				kycSvc.SetWebhookSecret(idv.ProviderPlaid, secret)
				whHandler.RegisterProvider(webhook.WebhookConfig{
					Provider:        idv.ProviderPlaid,
					Secret:          secret,
					SignatureHeader: "Plaid-Verification",
				}, makeKYCProcessor(kycSvc, idv.ProviderPlaid))
			}
		}
	}

	// Set default provider
	kycSvc.SetDefault(defaultProvider)
}

// makeKYCProcessor returns a webhook.ProcessorFunc that delegates to the KYC service.
func makeKYCProcessor(kycSvc *kyc.Service, provider string) webhook.ProcessorFunc {
	return func(_ string, body []byte, headers map[string]string) (string, error) {
		event, err := kycSvc.HandleWebhook(provider, body, headers)
		if err != nil {
			return "", err
		}
		return event.VerificationID, nil
	}
}

// installDefaultRules adds sensible AML monitoring rules.
func installDefaultRules(m *aml.MonitoringService) {
	m.AddRule(aml.Rule{
		ID:              "default_single_10k",
		Type:            aml.RuleSingleAmount,
		Description:     "Flag single transactions >= $10,000",
		Enabled:         true,
		ThresholdAmount: 10000,
		Currency:        "USD",
		Severity:        aml.SeverityMedium,
	})
	m.AddRule(aml.Rule{
		ID:              "default_daily_25k",
		Type:            aml.RuleDailyAggregate,
		Description:     "Flag daily aggregate >= $25,000",
		Enabled:         true,
		ThresholdAmount: 25000,
		Currency:        "USD",
		Severity:        aml.SeverityHigh,
	})
	m.AddRule(aml.Rule{
		ID:                   "default_structuring",
		Type:                 aml.RuleStructuring,
		Description:          "Detect potential structuring around $10,000 CTR threshold",
		Enabled:              true,
		StructuringThreshold: 10000,
		StructuringMargin:    1000,
		StructuringMinCount:  3,
		Severity:             aml.SeverityCritical,
	})
	m.AddRule(aml.Rule{
		ID:                "default_velocity",
		Type:              aml.RuleVelocity,
		Description:       "Flag accounts with 20+ transactions per hour",
		Enabled:           true,
		MaxCount:          20,
		Window:            time.Hour,
		Severity:          aml.SeverityMedium,
	})
}
