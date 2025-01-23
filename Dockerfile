FROM golang:1.22-alpine

WORKDIR /app

# Instalar dependências de build
RUN apk add --no-cache gcc musl-dev

# Copiar arquivos do módulo Go
COPY go.mod go.sum ./
RUN go mod download

# Copiar o código fonte
COPY . .

# Compilar a aplicação
RUN go build -o main ./cmd/api

# Expor a porta
EXPOSE 8080

# Comando para executar a aplicação
CMD ["./main"]