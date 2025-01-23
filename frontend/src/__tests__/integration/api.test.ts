import { authApi, emailsApi } from '../../lib/api';

describe('Testes de Integração API', () => {
  const testUser = {
    name: 'Usuário Teste',
    email: `test${Date.now()}@example.com`,
    password: 'senha123',
  };

  let authToken: string;

  describe('Autenticação', () => {
    test('deve registrar um novo usuário', async () => {
      const response = await authApi.register(testUser);
      
      expect(response).toHaveProperty('token');
      expect(response).toHaveProperty('user');
      expect(response.user).toHaveProperty('id');
      expect(response.user.email).toBe(testUser.email);
      expect(response.user.name).toBe(testUser.name);
      
      authToken = response.token;
    });

    test('deve fazer login com o usuário criado', async () => {
      const response = await authApi.login({
        email: testUser.email,
        password: testUser.password,
      });

      expect(response).toHaveProperty('token');
      expect(response).toHaveProperty('user');
      expect(response.user.email).toBe(testUser.email);
      
      authToken = response.token;
    });
  });

  describe('Emails', () => {
    test('deve buscar emails sem filtros', async () => {
      const emails = await emailsApi.getEmails();
      expect(Array.isArray(emails)).toBe(true);
    });

    test('deve buscar emails com filtro de prioridade', async () => {
      const emails = await emailsApi.getEmails({ priority: 'high' });
      expect(Array.isArray(emails)).toBe(true);
      emails.forEach(email => {
        expect(email.priority).toBe('high');
      });
    });

    test('deve buscar emails com filtro de categoria', async () => {
      const emails = await emailsApi.getEmails({ category: 'work' });
      expect(Array.isArray(emails)).toBe(true);
      emails.forEach(email => {
        expect(email.category).toBe('work');
      });
    });

    test('deve buscar emails com múltiplos filtros', async () => {
      const emails = await emailsApi.getEmails({
        priority: 'high',
        category: 'work',
        status: 'unread'
      });
      expect(Array.isArray(emails)).toBe(true);
    });
  });

  describe('Logout', () => {
    test('deve fazer logout do usuário', async () => {
      await expect(authApi.logout()).resolves.not.toThrow();
    });

    test('deve falhar ao tentar buscar emails após logout', async () => {
      await expect(emailsApi.getEmails()).rejects.toThrow();
    });
  });
});