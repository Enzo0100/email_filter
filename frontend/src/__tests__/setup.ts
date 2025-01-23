// Configurar variáveis de ambiente para teste
process.env.NEXT_PUBLIC_API_URL = 'http://localhost:8080';

// Aumentar timeout para requests de rede
jest.setTimeout(10000);

// Limpar todos os mocks após cada teste
afterEach(() => {
  jest.clearAllMocks();
});

// Configurar fetch global para Node.js
import fetch from 'node-fetch';
global.fetch = fetch as unknown as typeof global.fetch;
