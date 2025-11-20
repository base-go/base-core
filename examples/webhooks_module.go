package examples

import (
	"base/core/logger"
	"base/core/module"
	"base/core/router"
)

// WebhooksModule demonstrates how to use configurable middleware for webhook endpoints
type WebhooksModule struct {
	deps module.Dependencies
}

// NewWebhooksModule creates a new webhooks module
func NewWebhooksModule() *WebhooksModule {
	return &WebhooksModule{}
}

// Init initializes the webhooks module
func (m *WebhooksModule) Init(deps module.Dependencies) error {
	m.deps = deps
	return nil
}

// Migrate runs database migrations (none needed for this example)
func (m *WebhooksModule) Migrate() error {
	return nil
}

// Routes registers webhook routes with custom middleware configuration
func (m *WebhooksModule) Routes(router *router.RouterGroup) {
	// Create webhooks group
	webhooks := router.Group("/webhooks")
	
	// Stripe webhook endpoint
	webhooks.POST("/stripe", m.handleStripeWebhook)
	
	// GitHub webhook endpoint
	webhooks.POST("/github", m.handleGitHubWebhook)
	
	// PayPal webhook endpoint
	webhooks.POST("/paypal", m.handlePayPalWebhook)
	
	// Public webhook endpoint (no signature verification)
	webhooks.POST("/public", m.handlePublicWebhook)
}

// MiddlewareConfig returns middleware configuration overrides for webhook endpoints
func (m *WebhooksModule) MiddlewareConfig() *module.MiddlewareOverrides {
	return &module.MiddlewareOverrides{
		PathRules: map[string]module.MiddlewareSettings{
			// All webhook endpoints: disable API key and auth, enable signature verification
			"/api/webhooks/stripe": {
				APIKey: module.DisableAPIKey().APIKey,
				Auth:   module.DisableAuth().Auth,
				WebhookSignature: module.WebhookSignature(
					"stripe",
					"stripe-signature", 
					"STRIPE_WEBHOOK_SECRET",
				).WebhookSignature,
				RateLimit: module.CustomRateLimit(500, "1h").RateLimit,
			},
			
			"/api/webhooks/github": {
				APIKey: module.DisableAPIKey().APIKey,
				Auth:   module.DisableAuth().Auth,
				WebhookSignature: module.WebhookSignature(
					"github",
					"x-hub-signature-256",
					"GITHUB_WEBHOOK_SECRET",
				).WebhookSignature,
				RateLimit: module.CustomRateLimit(200, "1h").RateLimit,
			},
			
			"/api/webhooks/paypal": {
				APIKey: module.DisableAPIKey().APIKey,
				Auth:   module.DisableAuth().Auth,
				WebhookSignature: module.WebhookSignature(
					"paypal",
					"paypal-transmission-sig",
					"PAYPAL_WEBHOOK_SECRET",
				).WebhookSignature,
				RateLimit: module.CustomRateLimit(300, "1h").RateLimit,
			},
			
			// Public webhook: no security, higher rate limit
			"/api/webhooks/public": {
				APIKey:           module.DisableAPIKey().APIKey,
				Auth:             module.DisableAuth().Auth,
				WebhookSignature: nil, // No signature verification
				RateLimit:       module.CustomRateLimit(1000, "1h").RateLimit,
			},
		},
	}
}

// Webhook handlers

func (m *WebhooksModule) handleStripeWebhook(c *router.Context) error {
	m.deps.Logger.Info("Received Stripe webhook")
	
	// Parse Stripe webhook payload
	var payload map[string]interface{}
	if err := c.BindJSON(&payload); err != nil {
		return c.JSON(400, map[string]string{"error": "Invalid JSON"})
	}
	
	// Process Stripe event
	eventType, ok := payload["type"].(string)
	if !ok {
		return c.JSON(400, map[string]string{"error": "Missing event type"})
	}
	
	m.deps.Logger.Info("Processing Stripe event", 
		logger.String("type", eventType))
	
	// Handle different Stripe events
	switch eventType {
	case "payment_intent.succeeded":
		return m.handlePaymentSuccess(c, payload)
	case "payment_intent.payment_failed":
		return m.handlePaymentFailure(c, payload)
	default:
		m.deps.Logger.Info("Unhandled Stripe event type", 
			logger.String("type", eventType))
	}
	
	return c.JSON(200, map[string]string{"status": "received"})
}

func (m *WebhooksModule) handleGitHubWebhook(c *router.Context) error {
	m.deps.Logger.Info("Received GitHub webhook")
	
	// Get GitHub event type from header
	eventType := c.GetHeader("x-github-event")
	if eventType == "" {
		return c.JSON(400, map[string]string{"error": "Missing GitHub event type"})
	}
	
	var payload map[string]interface{}
	if err := c.BindJSON(&payload); err != nil {
		return c.JSON(400, map[string]string{"error": "Invalid JSON"})
	}
	
	m.deps.Logger.Info("Processing GitHub event",
		logger.String("type", eventType))
	
	// Handle different GitHub events
	switch eventType {
	case "push":
		return m.handleGitHubPush(c, payload)
	case "pull_request":
		return m.handleGitHubPullRequest(c, payload)
	case "issues":
		return m.handleGitHubIssue(c, payload)
	default:
		m.deps.Logger.Info("Unhandled GitHub event type",
			logger.String("type", eventType))
	}
	
	return c.JSON(200, map[string]string{"status": "received"})
}

func (m *WebhooksModule) handlePayPalWebhook(c *router.Context) error {
	m.deps.Logger.Info("Received PayPal webhook")
	
	var payload map[string]interface{}
	if err := c.BindJSON(&payload); err != nil {
		return c.JSON(400, map[string]string{"error": "Invalid JSON"})
	}
	
	// PayPal webhook processing
	eventType, ok := payload["event_type"].(string)
	if !ok {
		return c.JSON(400, map[string]string{"error": "Missing event type"})
	}
	
	m.deps.Logger.Info("Processing PayPal event",
		logger.String("type", eventType))
	
	return c.JSON(200, map[string]string{"status": "received"})
}

func (m *WebhooksModule) handlePublicWebhook(c *router.Context) error {
	m.deps.Logger.Info("Received public webhook")
	
	var payload map[string]interface{}
	if err := c.BindJSON(&payload); err != nil {
		return c.JSON(400, map[string]string{"error": "Invalid JSON"})
	}
	
	// Simple public webhook processing
	m.deps.Logger.Info("Processing public webhook",
		logger.Any("payload_keys", getKeys(payload)))
	
	return c.JSON(200, map[string]string{"status": "received"})
}

// Helper handlers for specific events

func (m *WebhooksModule) handlePaymentSuccess(c *router.Context, payload map[string]interface{}) error {
	m.deps.Logger.Info("Payment succeeded", logger.Any("payload", payload))
	
	// TODO: Update order status, send confirmation email, etc.
	
	return c.JSON(200, map[string]string{"status": "payment_processed"})
}

func (m *WebhooksModule) handlePaymentFailure(c *router.Context, payload map[string]interface{}) error {
	m.deps.Logger.Info("Payment failed", logger.Any("payload", payload))
	
	// TODO: Handle payment failure, notify user, etc.
	
	return c.JSON(200, map[string]string{"status": "payment_failure_processed"})
}

func (m *WebhooksModule) handleGitHubPush(c *router.Context, payload map[string]interface{}) error {
	m.deps.Logger.Info("GitHub push event", logger.Any("payload", payload))
	
	// TODO: Trigger CI/CD pipeline, update deployment, etc.
	
	return c.JSON(200, map[string]string{"status": "push_processed"})
}

func (m *WebhooksModule) handleGitHubPullRequest(c *router.Context, payload map[string]interface{}) error {
	m.deps.Logger.Info("GitHub pull request event", logger.Any("payload", payload))
	
	// TODO: Run tests, update PR status, etc.
	
	return c.JSON(200, map[string]string{"status": "pr_processed"})
}

func (m *WebhooksModule) handleGitHubIssue(c *router.Context, payload map[string]interface{}) error {
	m.deps.Logger.Info("GitHub issue event", logger.Any("payload", payload))
	
	// TODO: Create internal ticket, notify team, etc.
	
	return c.JSON(200, map[string]string{"status": "issue_processed"})
}

// Helper function to get keys from a map
func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
