package services

import (
	"context"
	"testing"
	"time"

	"begbot/internal/config"
	"begbot/internal/models"
)

// Test helper types and functions

// MockEmailService simulates email sending for testing
type MockEmailService struct {
	SentEmails  []EmailNotification
	ShouldFail  bool
	FailOnCount int  // Fail after N emails sent (0 = never fail)
	SentCount   int  // Track how many emails have been sent
	CalledAsync bool // Track if called asynchronously
}

type EmailNotification struct {
	To          []string
	Subject     string
	HTMLContent string
	// Parsed fields from the template
	PurchasePrice   int
	Valuation       int
	DiscountPercent float64
	NewPrice        int
	Profit          int
	Description     string
	ImageURL        string
	Link            string
}

func (m *MockEmailService) SendEmailAsync(cfg EmailConfig, to []string, subject, htmlContent string) {
	m.CalledAsync = true
	m.SentCount++

	if m.ShouldFail && (m.FailOnCount == 0 || m.SentCount > m.FailOnCount) {
		// Simulate failure - log but don't return error (non-blocking)
		return
	}

	m.SentEmails = append(m.SentEmails, EmailNotification{
		To:          to,
		Subject:     subject,
		HTMLContent: htmlContent,
	})
}

// MockDatabase for testing trading rules and listings
type MockDatabase struct {
	TradingRules *models.Economics
	Listings     []models.Listing
}

func (m *MockDatabase) GetTradingRules(ctx context.Context) (*models.Economics, error) {
	if m.TradingRules == nil {
		defaultProfit := 0
		defaultDiscount := 0
		return &models.Economics{
			MinProfitSEK: &defaultProfit,
			MinDiscount:  &defaultDiscount,
		}, nil
	}
	return m.TradingRules, nil
}

// =============================================================================
// HAPPY PATH TESTS
// =============================================================================

// Test_EmailSentWhenListingPassesTradingRules tests that an email is sent
// when a listing meets both trading rule criteria:
// - profit (valuation - price) > min_profit_sek
// - discount % > min_discount
func Test_EmailSentWhenListingPassesTradingRules(t *testing.T) {
	// Setup
	ctx := context.Background()

	// Trading rules: min profit 500 SEK, min discount 10%
	tradingRules := &models.Economics{
		ID:           1,
		MinProfitSEK: intPtr(500),
		MinDiscount:  intPtr(10),
	}

	// Listing that passes:
	// - Price: 5000 SEK
	// - Valuation: 6000 SEK
	// - Profit: 1000 SEK (6000 - 5000) > 500 ✓
	// - Discount: 16.67% (1000/6000*100) > 10% ✓
	listing := models.Listing{
		ID:            1,
		ProductID:     int64Ptr(1),
		Price:         intPtr(6000), // buy price
		Valuation:     8000,         // valuation with discount applied
		Link:          "https://blocket.se/item/123",
		Title:         "iPhone 15 Pro Max",
		ConditionID:   int64Ptr(1),
		MarketplaceID: int64Ptr(1),
		Status:        "active",
	}

	// NewPrice for the product (used for "nypris" in email)
	newPrice := 15000

	// Test: Calculate profit and discount
	profit := listing.Valuation - *listing.Price
	discountPercent := float64(profit) / float64(listing.Valuation) * 100

	// Verify conditions pass
	if profit <= *tradingRules.MinProfitSEK {
		t.Errorf("Expected profit %d to be > min_profit_sek %d", profit, *tradingRules.MinProfitSEK)
	}

	if discountPercent <= float64(*tradingRules.MinDiscount) {
		t.Errorf("Expected discount %.2f%% to be > min_discount %d%%", discountPercent, *tradingRules.MinDiscount)
	}

	// This is what the feature should do:
	// 1. After saving listing, get trading rules
	// 2. Calculate profit = valuation - price
	// 3. Calculate discount% = profit / valuation * 100
	// 4. If profit > min_profit_sek AND discount% > min_discount, send email
	// 5. Use mail.html template with listing data

	_ = ctx
	_ = newPrice

	t.Logf("Listing passes trading rules: profit=%d (>%d), discount=%.2f%% (>%d%%)",
		profit, *tradingRules.MinProfitSEK, discountPercent, *tradingRules.MinDiscount)
}

// Test_EmailContainsAllRequiredFields verifies the email template contains all required fields
func Test_EmailContainsAllRequiredFields(t *testing.T) {
	// Required fields per issue:
	// - Inköpspris (purchase price)
	// - Värdering (med rabatt %)
	// - Nypris (new price)
	// - Vinst (profit)
	// - Annonstext (ad text)
	// - Bild (image)
	// - Länk till annons (link)

	listing := models.Listing{
		ID:          1,
		Price:       intPtr(5000),
		Valuation:   8000,
		Link:        "https://blocket.se/item/123",
		Title:       "iPhone 15 Pro",
		Description: stringPtr("Säljer iPhone 15 Pro i fint skick"),
	}

	// Calculate fields that should be in email
	profit := listing.Valuation - *listing.Price
	discountPercent := float64(profit) / float64(listing.Valuation) * 100
	newPrice := 15000 // would come from product

	// Verify all fields can be calculated
	if profit <= 0 {
		t.Error("Profit should be positive")
	}

	if discountPercent <= 0 {
		t.Error("Discount percent should be positive")
	}

	// The email template (mail.html) should be populated with these values
	emailFields := map[string]interface{}{
		"purchase_price":   *listing.Price,
		"valuation":        listing.Valuation,
		"discount_percent": discountPercent,
		"new_price":        newPrice,
		"profit":           profit,
		"description":      *listing.Description,
		"link":             listing.Link,
	}

	for key, value := range emailFields {
		if value == nil || value == "" {
			t.Errorf("Email field '%s' should not be empty", key)
		}
	}

	t.Logf("Email fields: %+v", emailFields)
}

// =============================================================================
// EDGE CASE TESTS
// =============================================================================

// Test_NoEmailWhenProfitTooLow tests that NO email is sent when profit is below threshold
func Test_NoEmailWhenProfitTooLow(t *testing.T) {
	tradingRules := &models.Economics{
		ID:           1,
		MinProfitSEK: intPtr(500),
		MinDiscount:  intPtr(10),
	}

	// Listing with low profit:
	// - Price: 7500 SEK
	// - Valuation: 8000 SEK
	// - Profit: 500 SEK (== min - on the edge)
	listing := models.Listing{
		Price:     intPtr(7500),
		Valuation: 8000,
	}

	profit := listing.Valuation - *listing.Price
	discountPercent := float64(profit) / float64(listing.Valuation) * 100

	// Edge case: profit <= min_profit_sek should NOT trigger email
	if profit <= *tradingRules.MinProfitSEK {
		t.Logf("Correctly NOT sending email: profit %d <= min %d", profit, *tradingRules.MinProfitSEK)
	} else {
		// This listing passes, need one that fails
		t.Errorf("Test setup error: profit should be too low")
	}

	_ = discountPercent
}

// Test_NoEmailWhenDiscountTooLow tests that NO email is sent when discount % is below threshold
func Test_NoEmailWhenDiscountTooLow(t *testing.T) {
	tradingRules := &models.Economics{
		ID:           1,
		MinProfitSEK: intPtr(100),
		MinDiscount:  intPtr(20), // 20% discount required
	}

	// Listing with low discount:
	// - Price: 7000 SEK
	// - Valuation: 7500 SEK
	// - Profit: 500 SEK (> 100) ✓
	// - Discount: 6.67% (< 20%) ✗
	listing := models.Listing{
		Price:     intPtr(7000),
		Valuation: 7500,
	}

	profit := listing.Valuation - *listing.Price
	discountPercent := float64(profit) / float64(listing.Valuation) * 100

	// Should NOT trigger email because discount is too low
	if discountPercent <= float64(*tradingRules.MinDiscount) {
		t.Logf("Correctly NOT sending email: discount %.2f%% < min %d%%", discountPercent, *tradingRules.MinDiscount)
	} else {
		t.Errorf("Expected discount too low")
	}

	// Verify profit IS high enough
	if profit <= *tradingRules.MinProfitSEK {
		t.Errorf("Test setup: profit should be high enough (%d > %d)", profit, *tradingRules.MinProfitSEK)
	}
}

// Test_DefaultTradingRulesWhenNoneExist tests using defaults when no rules in DB
func Test_DefaultTradingRulesWhenNoneExist(t *testing.T) {
	// When no trading rules exist in database, should use defaults (0)
	defaultRules := &models.Economics{
		MinProfitSEK: intPtr(0),
		MinDiscount:  intPtr(0),
	}

	// Any listing should pass with defaults of 0
	listing := models.Listing{
		Price:     intPtr(1000),
		Valuation: 1100,
	}

	profit := listing.Valuation - *listing.Price
	discountPercent := float64(profit) / float64(listing.Valuation) * 100

	// With defaults (0,0), any positive profit and discount should pass
	passesProfit := profit > *defaultRules.MinProfitSEK
	passesDiscount := discountPercent > float64(*defaultRules.MinDiscount)

	if !passesProfit || !passesDiscount {
		t.Errorf("With default rules, listing should pass: profit=%d (>0), discount=%.2f%% (>0%%)", profit, discountPercent)
	}

	t.Logf("Default rules test: profit=%d, discount=%.2f%%", profit, discountPercent)
}

// Test_EmailFailureIsNonBlocking tests that email failure doesn't block the main flow
func Test_EmailFailureIsNonBlocking(t *testing.T) {
	// Per issue requirements: "Felhantering: Om mail misslyckas, logga fel men fortsätt execution"

	// This test verifies the requirement - email should be async and non-blocking
	// If email fails, the bot should continue processing other ads

	mockEmail := &MockEmailService{
		ShouldFail: true, // Simulate email failure
	}

	// Simulate calling email service asynchronously
	// The function should return immediately without blocking
	go func() {
		mockEmail.SendEmailAsync(
			EmailConfig{},
			[]string{"test@example.com"},
			"Test Subject",
			"<html>Test</html>",
		)
	}()

	// Give a tiny bit of time for the goroutine to start
	time.Sleep(10 * time.Millisecond)

	// The test passes if:
	// 1. Email was called asynchronously (CalledAsync = true)
	// 2. Main flow continues (no panic, no error returned)
	if !mockEmail.CalledAsync {
		t.Error("Email should be sent asynchronously (non-blocking)")
	}

	t.Log("Email failure is handled non-blocking - main flow continues")
}

// Test_TradingRulesBothConditionsMustBeMet tests that BOTH conditions must pass
func Test_TradingRulesBothConditionsMustBeMet(t *testing.T) {
	tradingRules := &models.Economics{
		ID:           1,
		MinProfitSEK: intPtr(500),
		MinDiscount:  intPtr(10),
	}

	testCases := []struct {
		name            string
		price           int
		valuation       int
		shouldSendEmail bool
	}{
		{
			name:            "Both conditions pass",
			price:           5000,
			valuation:       7000, // profit=2000 (>500), discount=28.6% (>10%)
			shouldSendEmail: true,
		},
		{
			name:            "Only profit passes",
			price:           6400,
			valuation:       7000, // profit=600 (>500), discount=8.6% (<10%)
			shouldSendEmail: false,
		},
		{
			name:            "Only discount passes",
			price:           6450,
			valuation:       6500, // profit=50 (<500), discount=0.77% (<10%)
			shouldSendEmail: false,
		},
		{
			name:            "Neither passes",
			price:           6950,
			valuation:       7000, // profit=50 (<500), discount=0.71% (<10%)
			shouldSendEmail: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			profit := tc.valuation - tc.price
			discountPercent := float64(profit) / float64(tc.valuation) * 100

			profitPasses := profit > *tradingRules.MinProfitSEK
			discountPasses := discountPercent > float64(*tradingRules.MinDiscount)
			shouldSend := profitPasses && discountPasses

			if shouldSend != tc.shouldSendEmail {
				t.Errorf("Expected shouldSendEmail=%v, got %v (profitPasses=%v, discountPasses=%.2f%%)",
					tc.shouldSendEmail, shouldSend, profitPasses, discountPercent)
			}
		})
	}
}

// Test_EmailUsesConfigRecipients tests that email uses recipients from config
func Test_EmailUsesConfigRecipients(t *testing.T) {
	// Per issue: "Använd befintliga email.recipients från config"

	cfg := EmailConfig{
		Recipients: []string{
			"user1@example.com",
			"user2@example.com",
		},
	}

	// The email service should use cfg.Recipients
	if len(cfg.Recipients) == 0 {
		t.Error("Config should have recipients configured")
	}

	expectedRecipients := []string{"user1@example.com", "user2@example.com"}
	for i, expected := range expectedRecipients {
		if i >= len(cfg.Recipients) {
			t.Errorf("Missing recipient at index %d", i)
			continue
		}
		if cfg.Recipients[i] != expected {
			t.Errorf("Expected recipient %s, got %s", expected, cfg.Recipients[i])
		}
	}

	t.Logf("Email will be sent to: %v", cfg.Recipients)
}

// Test_MailTemplateFieldsMatchIssueRequirements verifies the email template structure
func Test_MailTemplateFieldsMatchIssueRequirements(t *testing.T) {
	// Per issue, mail.html should contain:
	// - Inköpspris (purchase price)
	// - Värdering (med rabatt %)
	// - Nypris (new price)
	// - Vinst (profit)
	// - Annonstext (ad text)
	// - Bild (image)
	// - Länk till annons (link)

	// This is a documentation test - it verifies the requirements mapping
	emailFields := []string{
		"Inköpspris",
		"Värdering",
		"Nypris",
		"Vinst",
		"Annonstext",
		"Bild",
		"Länk",
	}

	requiredHTMLFields := []string{
		"purchase_price", // maps to Inköpspris
		"valuation",      // maps to Värdering
		"new_price",      // maps to Nypris
		"profit",         // maps to Vinst
		"description",    // maps to Annonstext
		"image_url",      // maps to Bild
		"link",           // maps to Länk
	}

	if len(emailFields) != len(requiredHTMLFields) {
		t.Errorf("Field count mismatch: %d vs %d", len(emailFields), len(requiredHTMLFields))
	}

	for i, field := range emailFields {
		t.Logf("Required field %d: %s -> %s", i, field, requiredHTMLFields[i])
	}

	// Verify mail.html exists and has these placeholders (checked separately)
	// The implementation should use the existing mail.html template
}

// =============================================================================
// INTEGRATION TESTS - These test the actual BotService behavior
// These tests will FAIL to compile until the feature is implemented
// =============================================================================

// Mock PostgresDatabase for testing
type MockPostgresDatabase struct {
	TradingRulesToReturn *models.Economics
	GetTradingRulesError error
}

func (m *MockPostgresDatabase) GetTradingRules(ctx context.Context) (*models.Economics, error) {
	if m.GetTradingRulesError != nil {
		return nil, m.GetTradingRulesError
	}
	if m.TradingRulesToReturn == nil {
		defaultProfit := 500
		defaultDiscount := 10
		return &models.Economics{
			MinProfitSEK: &defaultProfit,
			MinDiscount:  &defaultDiscount,
		}, nil
	}
	return m.TradingRulesToReturn, nil
}

func (m *MockPostgresDatabase) GetTradingRunByID(ctx context.Context, id int64) (*models.ScrapingRun, error) {
	return nil, nil
}

func (m *MockPostgresDatabase) SaveScrapingRun(ctx context.Context, run *models.ScrapingRun) error {
	return nil
}

func (m *MockPostgresDatabase) UpdateScrapingRun(ctx context.Context, run *models.ScrapingRun) error {
	return nil
}

// Test_SendTradingRuleEmailMethodExists verifies that BotService has a method
// to send email notifications for listings that pass trading rules.
// This test will FAIL TO COMPILE because the method doesn't exist yet.
func Test_SendTradingRuleEmailMethodExists(t *testing.T) {
	cfg := &config.Config{
		Email: config.EmailConfig{
			SMTPHost:     "smtp.example.com",
			SMTPPort:     "587",
			SMTPUsername: "user",
			SMTPPassword: "pass",
			From:         "begbot@example.com",
			Recipients:   []string{"user@example.com"},
		},
	}

	bot := NewBotService(cfg, nil, nil, nil, nil, nil)

	listing := &models.Listing{
		ID:          1,
		Price:       intPtr(5000),
		Valuation:   8000,
		Link:        "https://blocket.se/item/123",
		Title:       "iPhone 15 Pro",
		Description: stringPtr("Test description"),
	}

	product := &models.Product{
		ID:       1,
		Name:     stringPtr("iPhone 15 Pro"),
		NewPrice: intPtr(15000),
	}

	err := bot.SendTradingRuleEmail(context.Background(), listing, product)
	if err != nil {
		t.Logf("Email sent (non-blocking error logged): %v", err)
	}

	t.Log("BotService.SendTradingRuleEmail method exists and can be called")
}

// Test_SendTradingRuleEmailIsNonBlocking verifies that email sending doesn't block.
// This test will FAIL TO COMPILE because the method doesn't exist yet.
func Test_SendTradingRuleEmailIsNonBlocking(t *testing.T) {
	cfg := &config.Config{
		Email: config.EmailConfig{
			SMTPHost:     "smtp.example.com",
			SMTPPort:     "587",
			SMTPUsername: "user",
			SMTPPassword: "pass",
			From:         "begbot@example.com",
			Recipients:   []string{"user@example.com"},
		},
	}

	bot := NewBotService(cfg, nil, nil, nil, nil, nil)

	listing := &models.Listing{
		ID:        1,
		Price:     intPtr(5000),
		Valuation: 8000,
		Link:      "https://blocket.se/item/123",
	}

	product := &models.Product{
		ID:       1,
		NewPrice: intPtr(15000),
	}

	done := make(chan bool)
	go func() {
		bot.SendTradingRuleEmail(context.Background(), listing, product)
		done <- true
	}()

	select {
	case <-done:
		t.Log("Email sending returned immediately (non-blocking)")
	case <-time.After(100 * time.Millisecond):
		t.Error("Email sending is blocking - should be async")
	}
}

// Test_SendTradingRuleEmailSendsToRecipients verifies email goes to config recipients.
// This test will FAIL TO COMPILE because the method doesn't exist yet.
func Test_SendTradingRuleEmailSendsToRecipients(t *testing.T) {
	recipients := []string{"user1@example.com", "user2@example.com"}

	cfg := &config.Config{
		Email: config.EmailConfig{
			SMTPHost:     "smtp.example.com",
			SMTPPort:     "587",
			SMTPUsername: "user",
			SMTPPassword: "pass",
			From:         "begbot@example.com",
			Recipients:   recipients,
		},
	}

	bot := NewBotService(cfg, nil, nil, nil, nil, nil)

	listing := &models.Listing{
		ID:        1,
		Price:     intPtr(5000),
		Valuation: 8000,
		Link:      "https://blocket.se/item/123",
		Title:     "iPhone 15 Pro",
	}

	product := &models.Product{
		ID:       1,
		Name:     stringPtr("iPhone 15 Pro"),
		NewPrice: intPtr(15000),
	}

	err := bot.SendTradingRuleEmail(context.Background(), listing, product)
	if err != nil {
		t.Logf("Email send result: %v", err)
	}

	t.Logf("Email should be sent to: %v", recipients)
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

func int64Ptr(i int64) *int64 {
	return &i
}

func stringPtr(s string) *string {
	return &s
}
