FROM node:20-alpine

WORKDIR /app

# Copiar arquivos de configuração
COPY package*.json ./
COPY tsconfig.json ./
COPY next.config.ts ./
COPY postcss.config.mjs ./
COPY tailwind.config.ts ./
COPY eslint.config.mjs ./
COPY prisma ./prisma/

# Instalar dependências
RUN npm ci

# Gerar cliente Prisma
RUN npx prisma generate

# Copiar código fonte
COPY src ./src
COPY public ./public

# Construir a aplicação
RUN npm run build

# Expor a porta
EXPOSE 3000

# Iniciar em modo de produção
CMD ["sh", "-c", "npx prisma migrate deploy && npm start"]
