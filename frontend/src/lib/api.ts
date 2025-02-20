const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

interface Email {
  id: string;
  subject: string;
  from: string;
  content: string;
  priority: string;
  category: string;
  labels: string[];
  processedAt: string;
  createdAt: string;
}

interface EmailFilters {
  priority?: string;
  category?: string;
  status?: string;
}

interface LoginCredentials {
  email: string;
  password: string;
}

interface RegisterData {
  name: string;
  email: string;
  password: string;
  companyName: string;
}

interface AuthResponse {
  token: string;
  user: {
    id: string;
    name: string;
    email: string;
  };
}

export const authApi = {
  async login(credentials: LoginCredentials): Promise<AuthResponse> {
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/auth/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json',
        },
        body: JSON.stringify(credentials),
        credentials: 'include',
      });

      const contentType = response.headers.get('content-type');
      console.debug('Login response:', {
        status: response.status,
        statusText: response.statusText,
        contentType,
        headers: Object.fromEntries(response.headers.entries()),
      });

      const responseText = await response.text();
      console.debug('Response body:', responseText);

      if (!response.ok) {
        let errorMessage = 'Falha na autenticação';
        
        if (responseText) {
          try {
            if (contentType?.includes('application/json')) {
              const error = JSON.parse(responseText);
              errorMessage = error.error || error.message || errorMessage;
            } else {
              errorMessage = responseText;
            }
          } catch (parseError) {
            console.error('Erro ao processar resposta JSON:', parseError);
          }
        }
        
        throw new Error(errorMessage);
      }

      try {
        const data = JSON.parse(responseText);
        if (!data.token) {
          throw new Error('Token não encontrado na resposta');
        }
        
        // Armazenar o token
        localStorage.setItem('token', data.token);
        
        return data;
      } catch (parseError) {
        console.error('Erro ao processar resposta de sucesso:', parseError);
        throw new Error('Erro ao processar resposta do servidor');
      }
    } catch (error) {
      console.error('Erro completo na requisição:', error);
      throw error;
    }
  },

  async register(data: RegisterData): Promise<AuthResponse> {
    const response = await fetch(`${API_BASE_URL}/api/v1/auth/register`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        name: data.name,
        email: data.email,
        password: data.password,
        tenant: {
          name: data.companyName,
          plan: 'free'
        }
      }),
      credentials: 'include',
    });

    if (!response.ok) {
      const contentType = response.headers.get('content-type');
      console.debug('Response status:', response.status);
      console.debug('Content-Type:', contentType);
      
      try {
        const responseText = await response.text();
        console.debug('Response body:', responseText);
        
        if (contentType?.includes('application/json')) {
          const error = JSON.parse(responseText);
          throw new Error(error.error || error.message || 'Falha no registro');
        }
        
        throw new Error(`Falha no registro: ${responseText}`);
      } catch (parseError) {
        console.error('Erro ao processar resposta:', parseError);
        throw new Error('Falha no registro: erro ao processar resposta do servidor');
      }
    }

    return response.json();
  },

  async logout(): Promise<void> {
    const response = await fetch(`${API_BASE_URL}/api/v1/auth/logout`, {
      method: 'POST',
      credentials: 'include',
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || 'Falha ao fazer logout');
    }
  },
};

export const emailsApi = {
  async getEmails(filters: EmailFilters = {}): Promise<Email[]> {
    const queryParams = new URLSearchParams();
    
    if (filters.priority) queryParams.append('priority', filters.priority);
    if (filters.category) queryParams.append('category', filters.category);
    if (filters.status) queryParams.append('status', filters.status);
    
    const queryString = queryParams.toString();
    const url = `${API_BASE_URL}/api/v1/emails${queryString ? `?${queryString}` : ''}`;
    
    try {
      const response = await fetch(url, {
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
        credentials: 'include',
      });

      if (!response.ok) {
        if (response.status === 401) {
          throw new Error('Sessão expirada. Por favor, faça login novamente.');
        }
        if (response.status === 403) {
          throw new Error('Você não tem permissão para acessar estes emails.');
        }
        if (response.status === 404) {
          throw new Error('Nenhum email encontrado.');
        }
        if (response.status >= 500) {
          throw new Error('Erro no servidor. Por favor, tente novamente em alguns minutos.');
        }
        
        const errorData = await response.json();
        throw new Error(errorData.message || 'Falha ao buscar emails');
      }

      return await response.json();
    } catch (error) {
      console.error('Erro ao buscar emails:', error);
      if (error instanceof Error) {
        throw error;
      }
      throw new Error('Erro ao conectar com o servidor. Verifique sua conexão.');
    }
  },
};
