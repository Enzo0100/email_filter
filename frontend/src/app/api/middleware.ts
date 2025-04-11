import { NextRequest, NextResponse } from 'next/server';
import jwt from 'jsonwebtoken';
import config from '@/lib/config';

export interface AuthenticatedRequest extends NextRequest {
  user?: {
    userId: string;
    tenantId: string;
    email: string;
    role: string;
  };
}

export async function authenticateRequest(
  request: AuthenticatedRequest
): Promise<NextResponse | null> {
  try {
    const token = request.headers.get('Authorization')?.replace('Bearer ', '');

    if (!token) {
      return NextResponse.json(
        { error: 'Token não fornecido' },
        { status: 401 }
      );
    }

    const decoded = jwt.verify(token, config.auth.jwtSecret) as {
      userId: string;
      tenantId: string;
      email: string;
      role: string;
    };

    request.user = decoded;
    return null;
  } catch (error) {
    return NextResponse.json(
      { error: 'Token inválido' },
      { status: 401 }
    );
  }
}

export async function validateTenant(
  request: AuthenticatedRequest
): Promise<NextResponse | null> {
  const tenantId = request.headers.get('X-Tenant-ID');

  if (!tenantId) {
    return NextResponse.json(
      { error: 'Tenant ID não fornecido' },
      { status: 400 }
    );
  }

  if (request.user?.tenantId !== tenantId) {
    return NextResponse.json(
      { error: 'Tenant ID inválido' },
      { status: 403 }
    );
  }

  return null;
}
