package services

import (
	"context"
	"time"

	"github.com/enzo010/email-filter/internal/application/services/nlp"
	"github.com/enzo010/email-filter/internal/domain/entities"
)

// EmailClassifier serviço responsável pela classificação de emails
type EmailClassifier struct {
	nlpModel *nlp.Model
}

// ClassificationResult resultado da classificação de um email
type ClassificationResult struct {
	Priority       entities.Priority `json:"priority"`
	Category       string            `json:"category"`
	Labels         []string          `json:"labels"`
	Confidence     float64           `json:"confidence"`
	SuggestedTasks []entities.Task   `json:"suggested_tasks"`
}

// ClassifyEmail classifica um email usando NLP
func (ec *EmailClassifier) ClassifyEmail(ctx context.Context, email *entities.Email) (*ClassificationResult, error) {
	// Preparar texto para análise
	text := email.Subject + "\n" + email.Content
	if err := ec.nlpModel.AnalyzeText(text); err != nil {
		return nil, err
	}

	// Classificar email usando o modelo NLP
	result := &ClassificationResult{
		Priority:       ec.nlpModel.ClassifyPriority(email),
		Category:       ec.nlpModel.ClassifyCategory(email),
		Labels:         ec.nlpModel.ExtractLabels(email),
		Confidence:     ec.nlpModel.AnalyzeConfidence(email),
		SuggestedTasks: ec.nlpModel.ExtractTasks(email),
	}

	// Atualizar timestamps das tarefas sugeridas
	now := time.Now()
	for i := range result.SuggestedTasks {
		result.SuggestedTasks[i].CreatedAt = now
		result.SuggestedTasks[i].UpdatedAt = now
	}

	return result, nil
}

// NewEmailClassifier cria uma nova instância do classificador
func NewEmailClassifier() *EmailClassifier {
	return &EmailClassifier{
		nlpModel: nlp.NewModel(),
	}
}
