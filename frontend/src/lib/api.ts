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
    const response = await fetch(`${API_BASE_URL}/api/v1/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(credentials),
      credentials: 'include',
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || 'Falha na autenticação');
    }

    return response.json();
  },

  async register(data: RegisterData): Promise<AuthResponse> {
    const response = await fetch(`${API_BASE_URL}/api/v1/auth/register`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
      credentials: 'include',
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || 'Falha no registro');
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
        },
        credentials: 'include',
      });

      if (!response.ok) {
        throw new Error('Falha ao buscar emails');
      }

      return await response.json();
    } catch (error) {
      console.error('Erro ao buscar emails:', error);
      throw error;
    }
  },
};