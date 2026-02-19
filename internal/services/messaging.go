package services

import (
	"context"
	"fmt"

	"begbot/internal/config"
	"begbot/internal/db"
	"begbot/internal/models"
)

type MessagingService struct {
	cfg        *config.Config
	db         *db.Postgres
	llmService *LLMService
}

func NewMessagingService(cfg *config.Config, database *db.Postgres, llmService *LLMService) *MessagingService {
	return &MessagingService{
		cfg:        cfg,
		db:         database,
		llmService: llmService,
	}
}

type MessageGenerationInput struct {
	ListingTitle        string
	ListingDescription  string
	ListingPrice        int
	Valuation           int
	ConversationHistory []models.Message
	MessageType         string // "initial" or "reply"
}

func (s *MessagingService) GenerateMessage(ctx context.Context, input MessageGenerationInput) (string, error) {
	var prompt string

	if input.MessageType == "initial" {
		prompt = s.buildInitialMessagePrompt(input)
	} else {
		prompt = s.buildReplyMessagePrompt(input)
	}

	model := s.llmService.client.GetModel("GenerateMessage", s.llmService.defaultModel, s.llmService.models)

	content, err := s.llmService.client.Chat(ctx, model, prompt)
	if err != nil {
		return "", fmt.Errorf("LLM API error: %w", err)
	}

	return content, nil
}

func (s *MessagingService) buildInitialMessagePrompt(input MessageGenerationInput) string {
	prompt := fmt.Sprintf(`Du är en hjälpsam köpare som är intresserad av att köpa en vara på en svensk marknadsplats.

Annonsdetaljer:
- Titel: %s
- Pris: %d kr
- Beskrivning: %s

Din värdering: %d kr (vad du är villig att betala)

Skriv ett kort, vänligt och naturligt meddelande på svenska för att visa intresse för varan. 
Meddelandet ska:
- Vara naturligt och personligt
- Inte nämna din maxpris direkt i första meddelandet
- Vara kortfattat (max 2-3 meningar)
- Fråga om varan fortfarande är till salu
- Eventuellt ställa en relevant fråga om skick eller användning

Returnera ENDAST meddelandet, ingen extra text eller förklaring.`,
		input.ListingTitle,
		input.ListingPrice,
		input.ListingDescription,
		input.Valuation,
	)
	return prompt
}

func (s *MessagingService) buildReplyMessagePrompt(input MessageGenerationInput) string {
	conversationHistory := ""
	for _, msg := range input.ConversationHistory {
		sender := "Säljare"
		if msg.Direction == "outgoing" {
			sender = "Du"
		}
		conversationHistory += fmt.Sprintf("%s: %s\n", sender, msg.Content)
	}

	prompt := fmt.Sprintf(`Du är en köpare som förhandlar om att köpa en vara på en svensk marknadsplats.

Annonsdetaljer:
- Titel: %s
- Pris: %d kr
- Beskrivning: %s

Din värdering: %d kr (maxpris du är villig att betala)

Konversationshistorik:
%s

Baserat på konversationen ovan, skriv ett lämpligt svar på svenska.
Meddelandet ska:
- Vara naturligt och personligt
- Fortsätta förhandlingen på ett artigt sätt
- Inte överskrida din värdering (%d kr)
- Vara kortfattat (max 2-3 meningar)
- Anpassa dig till tonaliteten i konversationen

Returnera ENDAST meddelandet, ingen extra text eller förklaring.`,
		input.ListingTitle,
		input.ListingPrice,
		input.ListingDescription,
		input.Valuation,
		conversationHistory,
		input.Valuation,
	)
	return prompt
}

func (s *MessagingService) CreateConversation(ctx context.Context, listingID, marketplaceID int64) (*models.Conversation, error) {
	// Check if conversation already exists for this listing
	existing, err := s.db.GetConversationByListingID(ctx, listingID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing conversation: %w", err)
	}
	if existing != nil {
		return existing, nil
	}

	conv := &models.Conversation{
		ListingID:     listingID,
		MarketplaceID: marketplaceID,
		Status:        "active",
	}

	if err := s.db.CreateConversation(ctx, conv); err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	return conv, nil
}

func (s *MessagingService) GenerateInitialMessage(ctx context.Context, listingID int64) (*models.Message, error) {
	// Get listing details
	listing, err := s.db.GetListingByID(ctx, listingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get listing: %w", err)
	}
	if listing == nil {
		return nil, fmt.Errorf("listing not found")
	}

	// Get or create conversation
	conv, err := s.CreateConversation(ctx, listingID, *listing.MarketplaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	// Generate message using LLM
	description := ""
	if listing.Description != nil {
		description = *listing.Description
	}

	price := 0
	if listing.Price != nil {
		price = *listing.Price
	}

	input := MessageGenerationInput{
		ListingTitle:       listing.Title,
		ListingDescription: description,
		ListingPrice:       price,
		Valuation:          listing.Valuation,
		MessageType:        "initial",
	}

	content, err := s.GenerateMessage(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to generate message: %w", err)
	}

	// Create message in database
	msg := &models.Message{
		ConversationID: conv.ID,
		Direction:      "outgoing",
		Content:        content,
		Status:         "pending",
	}

	if err := s.db.CreateMessage(ctx, msg); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	return msg, nil
}

func (s *MessagingService) GenerateReplyMessage(ctx context.Context, conversationID int64) (*models.Message, error) {
	// Get conversation
	conv, err := s.db.GetConversationByID(ctx, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}
	if conv == nil {
		return nil, fmt.Errorf("conversation not found")
	}

	// Get listing details
	listing, err := s.db.GetListingByID(ctx, conv.ListingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get listing: %w", err)
	}
	if listing == nil {
		return nil, fmt.Errorf("listing not found")
	}

	// Get conversation history
	messages, err := s.db.GetMessagesByConversationID(ctx, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	// Generate message using LLM
	description := ""
	if listing.Description != nil {
		description = *listing.Description
	}

	price := 0
	if listing.Price != nil {
		price = *listing.Price
	}

	input := MessageGenerationInput{
		ListingTitle:        listing.Title,
		ListingDescription:  description,
		ListingPrice:        price,
		Valuation:           listing.Valuation,
		ConversationHistory: messages,
		MessageType:         "reply",
	}

	content, err := s.GenerateMessage(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to generate message: %w", err)
	}

	// Create message in database
	msg := &models.Message{
		ConversationID: conv.ID,
		Direction:      "outgoing",
		Content:        content,
		Status:         "pending",
	}

	if err := s.db.CreateMessage(ctx, msg); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	return msg, nil
}

func (s *MessagingService) ReceiveMessage(ctx context.Context, conversationID int64, content string) (*models.Message, error) {
	msg := &models.Message{
		ConversationID: conversationID,
		Direction:      "incoming",
		Content:        content,
		Status:         "received",
	}

	if err := s.db.CreateMessage(ctx, msg); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	return msg, nil
}
