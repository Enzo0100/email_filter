package nlp

import (
	"fmt"
	"strings"
	"time"

	"github.com/bbalet/stopwords"
	"github.com/enzo010/email-filter/internal/domain/entities"
	"github.com/jdkato/prose/v2"
)

// parseDateEntity tenta extrair uma data de um texto
func parseDateEntity(text string) (time.Time, error) {
	text = strings.ToLower(text)
	now := time.Now()

	// Padrões comuns em português
	patterns := map[string]time.Time{
		"hoje":           now,
		"amanhã":         now.AddDate(0, 0, 1),
		"próxima semana": now.AddDate(0, 0, 7),
		"próximo mês":    now.AddDate(0, 1, 0),
		"semana que vem": now.AddDate(0, 0, 7),
		"mês que vem":    now.AddDate(0, 1, 0),
	}

	// Verificar padrões diretos
	for pattern, date := range patterns {
		if strings.Contains(text, pattern) {
			return date, nil
		}
	}

	// Tentar formatos de data comuns
	formats := []string{
		"02/01/2006",
		"02-01-2006",
		"2006-01-02",
		"02/01/06",
		"02-01-06",
	}

	for _, format := range formats {
		if date, err := time.Parse(format, text); err == nil {
			return date, nil
		}
	}

	// Tentar interpretar expressões relativas
	if strings.Contains(text, "dia") {
		parts := strings.Fields(text)
		for i, part := range parts {
			if part == "dia" && i+1 < len(parts) {
				if day, err := time.Parse("2", parts[i+1]); err == nil {
					return time.Date(now.Year(), now.Month(), day.Day(), 0, 0, 0, 0, now.Location()), nil
				}
			}
		}
	}

	return time.Time{}, fmt.Errorf("não foi possível extrair data do texto: %s", text)
}

// Model representa o modelo NLP para classificação de emails
type Model struct {
	doc *prose.Document
}

// NewModel cria uma nova instância do modelo NLP
func NewModel() *Model {
	return &Model{}
}

// AnalyzeText prepara o texto para análise
func (m *Model) AnalyzeText(text string) error {
	// Remover stopwords e normalizar texto
	cleanText := stopwords.CleanString(text, "pt", true)

	// Criar documento para análise
	doc, err := prose.NewDocument(cleanText)
	if err != nil {
		return fmt.Errorf("erro ao criar documento para análise: %v", err)
	}
	m.doc = doc
	return nil
}

// ClassifyPriority determina a prioridade do email baseado em análise de sentimento e urgência
func (m *Model) ClassifyPriority(email *entities.Email) entities.Priority {
	urgentTerms := map[string]bool{
		"urgente": true, "importante": true, "crítico": true,
		"emergência": true, "imediato": true, "prazo": true,
		"deadline": true, "urgent": true, "asap": true,
	}

	// Combinar subject e conteúdo para análise
	text := strings.ToLower(email.Subject + " " + email.Content)

	// Contar termos de urgência
	urgencyScore := 0
	for term := range urgentTerms {
		if strings.Contains(text, term) {
			urgencyScore++
		}
	}

	// Analisar entidades e tokens para identificar datas próximas
	hasNearDate := false
	for _, ent := range m.doc.Entities() {
		if ent.Label == "DATE" || ent.Label == "TIME" {
			// Tentar extrair data da entidade
			date, err := parseDateEntity(ent.Text)
			if err == nil {
				// Verificar se a data está dentro dos próximos 3 dias
				if time.Until(date) <= 72*time.Hour && time.Until(date) > 0 {
					hasNearDate = true
					break
				}
			}
		}
	}

	// Determinar prioridade baseado nos scores
	if urgencyScore >= 2 || hasNearDate {
		return entities.PriorityHigh
	} else if urgencyScore == 1 {
		return entities.PriorityMedium
	}
	return entities.PriorityLow
}

// ClassifyCategory determina a categoria do email baseado em análise de tópicos
func (m *Model) ClassifyCategory(email *entities.Email) string {
	categories := map[string][]string{
		"financeiro": {"pagamento", "fatura", "cobrança", "orçamento", "invoice", "payment"},
		"suporte":    {"problema", "erro", "bug", "ajuda", "support", "help"},
		"comercial":  {"proposta", "venda", "cliente", "reunião", "meeting", "sales"},
		"rh":         {"férias", "contrato", "ponto", "vacation", "contract", "hr"},
		"ti":         {"sistema", "acesso", "senha", "system", "password", "access"},
	}

	text := strings.ToLower(email.Subject + " " + email.Content)
	categoryScores := make(map[string]int)

	// Calcular score para cada categoria
	for category, terms := range categories {
		for _, term := range terms {
			if strings.Contains(text, term) {
				categoryScores[category]++
			}
		}
	}

	// Encontrar categoria com maior score
	maxScore := 0
	bestCategory := "outros"
	for category, score := range categoryScores {
		if score > maxScore {
			maxScore = score
			bestCategory = category
		}
	}

	return bestCategory
}

// ExtractLabels extrai labels relevantes do email
func (m *Model) ExtractLabels(email *entities.Email) []string {
	var labels []string
	labelSet := make(map[string]bool)

	// Extrair entidades nomeadas
	for _, ent := range m.doc.Entities() {
		switch ent.Label {
		case "PERSON":
			labelSet["pessoa"] = true
		case "ORG":
			labelSet["organização"] = true
		case "GPE", "LOC":
			labelSet["local"] = true
		case "PRODUCT":
			labelSet["produto"] = true
		}
	}

	// Extrair tópicos baseados em tokens
	text := strings.ToLower(email.Content)
	topics := map[string][]string{
		"projeto":     {"projeto", "project", "desenvolvimento", "development"},
		"reunião":     {"reunião", "meeting", "agenda", "scheduling"},
		"documento":   {"documento", "document", "contrato", "contract"},
		"treinamento": {"treinamento", "training", "curso", "course"},
	}

	for topic, terms := range topics {
		for _, term := range terms {
			if strings.Contains(text, term) {
				labelSet[topic] = true
				break
			}
		}
	}

	// Converter set para slice
	for label := range labelSet {
		labels = append(labels, label)
	}

	return labels
}

// ExtractTasks identifica possíveis tarefas no email
func (m *Model) ExtractTasks(email *entities.Email) []entities.Task {
	var tasks []entities.Task

	// Padrões que indicam tarefas
	actionPatterns := []string{
		"por favor", "preciso", "necessário",
		"favor", "please", "need",
		"deve", "should", "must",
		"poderia", "could", "can you",
	}

	// Dividir conteúdo em sentenças
	sentences := m.doc.Sentences()

	for _, sent := range sentences {
		text := strings.ToLower(sent.Text)

		// Verificar se a sentença contém padrões de ação
		isTask := false
		for _, pattern := range actionPatterns {
			if strings.Contains(text, pattern) {
				isTask = true
				break
			}
		}

		if isTask {
			// Criar tarefa
			task := entities.Task{
				Description: sent.Text,
				DueDate:     time.Now().Add(24 * time.Hour), // Default 24h
				Priority:    entities.PriorityMedium,
				Status:      "pending",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			// Tentar identificar prazo na sentença
			for _, ent := range m.doc.Entities() {
				if ent.Label == "DATE" || ent.Label == "TIME" {
					// TODO: Implementar parsing de data
					break
				}
			}

			tasks = append(tasks, task)
		}
	}

	return tasks
}

// AnalyzeConfidence calcula a confiança da classificação baseado em múltiplos fatores
func (m *Model) AnalyzeConfidence(email *entities.Email) float64 {
	var confidence float64 = 0.0
	totalFactors := 0.0

	// Fator 1: Comprimento do conteúdo (emails muito curtos têm menor confiança)
	contentLength := len(email.Content)
	if contentLength > 500 {
		confidence += 1.0
	} else if contentLength > 200 {
		confidence += 0.8
	} else if contentLength > 100 {
		confidence += 0.6
	} else {
		confidence += 0.4
	}
	totalFactors++

	// Fator 2: Presença de entidades nomeadas
	entityCount := len(m.doc.Entities())
	if entityCount > 5 {
		confidence += 1.0
	} else if entityCount > 3 {
		confidence += 0.8
	} else if entityCount > 1 {
		confidence += 0.6
	} else {
		confidence += 0.4
	}
	totalFactors++

	// Fator 3: Força da classificação de categoria
	categories := map[string][]string{
		"financeiro": {"pagamento", "fatura", "cobrança", "orçamento", "invoice"},
		"suporte":    {"problema", "erro", "bug", "ajuda", "support"},
		"comercial":  {"proposta", "venda", "cliente", "reunião", "meeting"},
		"rh":         {"férias", "contrato", "ponto", "vacation", "hr"},
		"ti":         {"sistema", "acesso", "senha", "system", "password"},
	}

	text := strings.ToLower(email.Subject + " " + email.Content)
	maxCategoryScore := 0
	for _, terms := range categories {
		score := 0
		for _, term := range terms {
			if strings.Contains(text, term) {
				score++
			}
		}
		if score > maxCategoryScore {
			maxCategoryScore = score
		}
	}

	if maxCategoryScore > 3 {
		confidence += 1.0
	} else if maxCategoryScore > 2 {
		confidence += 0.8
	} else if maxCategoryScore > 1 {
		confidence += 0.6
	} else {
		confidence += 0.4
	}
	totalFactors++

	// Fator 4: Presença de datas e horários
	hasTimeEntities := false
	for _, ent := range m.doc.Entities() {
		if ent.Label == "DATE" || ent.Label == "TIME" {
			hasTimeEntities = true
			break
		}
	}
	if hasTimeEntities {
		confidence += 1.0
	} else {
		confidence += 0.5
	}
	totalFactors++

	// Calcular média ponderada final
	return confidence / totalFactors
}
