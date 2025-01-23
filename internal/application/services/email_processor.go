package services

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
	"github.com/enzo010/email-filter/internal/domain/entities"
)

// EmailProcessor responsável por processar emails da caixa de entrada
type EmailProcessor struct {
	imapClient      *client.Client
	emailClassifier *EmailClassifier
	emailRepo       entities.EmailRepository
	config          *EmailConfig
	tenantID        string
	userID          string
}

// EmailConfig configuração para conexão com servidor de email
type EmailConfig struct {
	Server   string
	Port     int
	Username string
	Password string
	Folder   string // Ex: "INBOX"
	SSL      bool
	TenantID string // Adicionado campo TenantID
	UserID   string // Adicionado campo UserID
}

// NewEmailProcessor cria uma nova instância do processador de emails
func NewEmailProcessor(config *EmailConfig, classifier *EmailClassifier, repo entities.EmailRepository) (*EmailProcessor, error) {
	// Construir string de conexão
	addr := fmt.Sprintf("%s:%d", config.Server, config.Port)

	// Conectar ao servidor IMAP
	var c *client.Client
	var err error
	if config.SSL {
		c, err = client.DialTLS(addr, nil)
	} else {
		c, err = client.Dial(addr)
	}

	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao servidor IMAP: %v", err)
	}

	// Login
	if err := c.Login(config.Username, config.Password); err != nil {
		return nil, fmt.Errorf("erro no login: %v", err)
	}

	return &EmailProcessor{
		imapClient:      c,
		emailClassifier: classifier,
		emailRepo:       repo,
		config:          config,
		tenantID:        config.TenantID,
		userID:          config.UserID,
	}, nil
}

// StartProcessing inicia o processamento de emails
func (ep *EmailProcessor) StartProcessing(ctx context.Context) error {
	// Selecionar pasta
	if _, err := ep.imapClient.Select(ep.config.Folder, false); err != nil {
		return fmt.Errorf("erro ao selecionar pasta: %v", err)
	}

	// Configurar critérios de busca (apenas emails não lidos)
	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{imap.SeenFlag}

	// Canal para receber atualizações
	updates := make(chan client.Update)
	ep.imapClient.Updates = updates

	// Goroutine para processar atualizações
	go func() {
		for {
			select {
			case update := <-updates:
				if _, ok := update.(*client.MailboxUpdate); ok {
					// Processar novos emails quando houver atualização da caixa
					ep.processNewEmails(ctx)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Processar emails existentes não lidos
	ep.processNewEmails(ctx)

	// Iniciar modo IDLE para receber notificações de novos emails
	done := make(chan error, 1)
	stop := make(chan struct{})

	go func() {
		done <- ep.imapClient.Idle(stop, nil)
	}()

	// Parar IDLE quando o contexto for cancelado
	go func() {
		<-ctx.Done()
		close(stop)
	}()

	return nil
}

// processNewEmails processa emails não lidos
func (ep *EmailProcessor) processNewEmails(ctx context.Context) error {
	// Buscar emails não lidos
	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{imap.SeenFlag}

	uids, err := ep.imapClient.Search(criteria)
	if err != nil {
		return fmt.Errorf("erro ao buscar emails: %v", err)
	}

	if len(uids) == 0 {
		return nil
	}

	// Criar sequência de UIDs
	seqset := new(imap.SeqSet)
	seqset.AddNum(uids...)

	// Buscar mensagens
	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)

	go func() {
		done <- ep.imapClient.Fetch(seqset, []imap.FetchItem{
			imap.FetchEnvelope,
			imap.FetchFlags,
			imap.FetchBody,
			imap.FetchBodyStructure,
		}, messages)
	}()

	// Processar mensagens
	for msg := range messages {
		if err := ep.processMessage(ctx, msg); err != nil {
			log.Printf("Erro ao processar mensagem %d: %v", msg.Uid, err)
		}
	}

	return <-done
}

// processMessage processa uma única mensagem
func (ep *EmailProcessor) processMessage(ctx context.Context, msg *imap.Message) error {
	// Extrair informações do email
	var body string
	var subject string
	var from string
	var to string

	if msg.Envelope != nil {
		subject = msg.Envelope.Subject
		if len(msg.Envelope.From) > 0 {
			from = msg.Envelope.From[0].Address()
		}
		if len(msg.Envelope.To) > 0 {
			to = msg.Envelope.To[0].Address()
		}
	}

	// Extrair corpo do email
	for _, part := range msg.Body {
		mr, err := mail.CreateReader(part)
		if err != nil {
			continue
		}

		for {
			p, err := mr.NextPart()
			if err != nil {
				break
			}

			switch p.Header.Get("Content-Type") {
			case "text/plain":
				buf := new(strings.Builder)
				if _, err := io.Copy(buf, p.Body); err != nil {
					continue
				}
				body = buf.String()
			}
		}
	}

	// Criar entidade de email
	email := &entities.Email{
		TenantID: ep.tenantID, // Definir TenantID
		UserID:   ep.userID,   // Definir UserID
		Subject:  subject,
		From:     from,
		To:       to,
		Content:  body,
	}

	// Classificar email
	result, err := ep.emailClassifier.ClassifyEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("erro ao classificar email: %v", err)
	}

	// Atualizar email com resultados
	email.Priority = result.Priority
	email.Category = result.Category
	email.Labels = result.Labels
	email.Tasks = result.SuggestedTasks
	email.ProcessedAt = time.Now()

	// Salvar no banco de dados
	if err := ep.emailRepo.Create(ctx, email); err != nil {
		return fmt.Errorf("erro ao salvar email: %v", err)
	}

	// Marcar como lido
	seqset := new(imap.SeqSet)
	seqset.AddNum(msg.Uid)
	item := imap.FormatFlagsOp(imap.AddFlags, true)
	flags := []interface{}{imap.SeenFlag}
	if err := ep.imapClient.UidStore(seqset, item, flags, nil); err != nil {
		return fmt.Errorf("erro ao marcar email como lido: %v", err)
	}

	return nil
}

// Close fecha a conexão com o servidor IMAP
func (ep *EmailProcessor) Close() error {
	return ep.imapClient.Logout()
}
