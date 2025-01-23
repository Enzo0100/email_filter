# Email Filter - Sistema Inteligente de Classificação de E-mails

Sistema SaaS multitenancy para classificação inteligente de e-mails e geração automática de tarefas usando processamento de linguagem natural (NLP).

## Funcionalidades

- 📧 Classificação automática de e-mails por prioridade
- 🏷️ Categorização inteligente (financeiro, pessoal, suporte, etc.)
- ✅ Extração automática de tarefas do conteúdo dos e-mails
- 🏢 Suporte a múltiplos tenants (organizações)
- 🔄 Integração com diversos provedores de e-mail
- 📊 Dashboard com métricas e análises
- 🔌 API REST para integrações

## Planos

- **Free**
  - Até 100 emails/mês
  - Categorização básica
  - 30 dias de histórico

- **Pro**
  - Emails ilimitados
  - Categorização avançada
  - Integração com Slack/Teams
  - 1 ano de histórico

- **Enterprise**
  - Todas as funcionalidades Pro
  - Customização de categorias
  - Integração com sistemas internos
  - Histórico ilimitado
  - Suporte prioritário

## Requisitos

- Go 1.21+
- PostgreSQL 14+
- Redis (opcional, para cache)
- Docker (opcional)

## Configuração do Ambiente

1. Clone o repositório:
```bash
git clone https://github.com/enzo010/email-filter.git
cd email-filter
```

2. Configure as variáveis de ambiente:
```bash
cp .env.example .env
# Edite o arquivo .env com suas configurações
```

3. Instale as dependências:
```bash
go mod download
```

4. Configure o banco de dados:
```bash
# Crie o banco de dados
createdb email_filter

# Execute as migrações
psql -d email_filter -f internal/infrastructure/database/migrations/001_initial_schema.sql
```

5. Inicie o servidor:
```bash
go run cmd/api/main.go
```

## Estrutura do Projeto

```
.
├── cmd/
│   └── api/              # Ponto de entrada da aplicação
├── internal/
│   ├── domain/          # Regras de negócio e entidades
│   │   └── entities/    # Definição das entidades
│   ├── application/     # Casos de uso da aplicação
│   │   └── services/    # Serviços da aplicação
│   └── infrastructure/  # Implementações concretas
│       └── database/    # Camada de banco de dados
├── pkg/                 # Bibliotecas compartilhadas
└── api/                 # Documentação da API
```

## Desenvolvimento

### Padrões de Código

- Seguimos a [Effective Go](https://golang.org/doc/effective_go)
- Utilizamos Clean Architecture
- Código e comentários em português
- Testes unitários para lógica de negócio

### Testes

```bash
# Executar todos os testes
go test ./...

# Executar testes com cobertura
go test -cover ./...
```

## API Documentation

A documentação da API está disponível em `/api/swagger.yaml`

## Contribuindo

1. Fork o projeto
2. Crie uma branch para sua feature (`git checkout -b feature/nova-feature`)
3. Commit suas mudanças (`git commit -m 'Adiciona nova feature'`)
4. Push para a branch (`git push origin feature/nova-feature`)
5. Abra um Pull Request

## Licença

Este projeto está licenciado sob a licença MIT - veja o arquivo [LICENSE](LICENSE) para detalhes.

## Contato

- Email: suporte@emailfilter.com.br
- Website: https://emailfilter.com.br