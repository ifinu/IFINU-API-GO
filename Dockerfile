# Stage 1: Build
FROM golang:1.22-alpine AS builder

# Informações do build
LABEL maintainer="IFINU <contato@ifinu.io>"
LABEL description="IFINU API - Sistema de Cobrança Online"

WORKDIR /app

# Instalar dependências de build
RUN apk add --no-cache git ca-certificates tzdata

# Copiar módulos Go e baixar dependências
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copiar código fonte
COPY . .

# Build do binário otimizado
# CGO_ENABLED=0 para binário estático
# -ldflags="-w -s" remove informações de debug (reduz tamanho)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a -installsuffix cgo \
    -ldflags="-w -s" \
    -o ifinu-api \
    ./cmd/api

# Stage 2: Runtime
FROM alpine:latest

# Instalar certificados SSL e timezone
RUN apk --no-cache add ca-certificates tzdata

# Definir timezone para São Paulo
ENV TZ=America/Sao_Paulo

WORKDIR /root/

# Copiar binário do stage de build
COPY --from=builder /app/ifinu-api .

# Copiar arquivo .env se existir (para produção use secrets)
COPY .env* ./

# Criar diretório de uploads
RUN mkdir -p uploads && chmod 755 uploads

# Expor porta
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Executar aplicação
CMD ["./ifinu-api"]
