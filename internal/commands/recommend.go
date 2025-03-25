package commands

import (
	"blockmind/internal/config"
	"blockmind/internal/crypto"
	"blockmind/internal/ia"
	"context"
	"strings"
)

type RecommendCommand struct {
	cfg *config.Config
}

func NewRecommendCommand(cfg *config.Config) *RecommendCommand {
	return &RecommendCommand{cfg: cfg}
}

func (c *RecommendCommand) Name() string {
	return "recommend"
}

func (c *RecommendCommand) Aliases() []string {
	return []string{"r", "recomendar"}
}

func (c *RecommendCommand) Description() string {
	return "Get a recommendation for a cryptocurrency"
}

func (c *RecommendCommand) Execute(ctx context.Context, args []string) (string, error) {
	if len(args) == 0 {
		return "Please specify a cryptocurrency (e.g., /recommend Bitcoin)", nil
	}

	cryptoName := strings.Join(args, " ")
	recommendation_data, err := crypto.GetCryptoRecommendation(cryptoName, c.cfg)
	if err != nil {
		return "", err
	}

	recommendation_data, err = crypto.GetSentimentAndHistoricalData(recommendation_data, cryptoName, c.cfg)

	recommendation, err := ia.GetInvestmentRecommendation(cryptoName, recommendation_data, c.cfg)
	if err != nil {
		return "", err
	}

	recommendation = recommendation + "\n\n" + "*This is not financial advice. Always do your own research.*"

	return recommendation, nil
}
