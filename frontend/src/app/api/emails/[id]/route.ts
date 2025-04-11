import { NextRequest, NextResponse } from 'next/server';
import { authenticateRequest, validateTenant, AuthenticatedRequest } from '../../middleware';
import { prisma } from '@/lib/prisma';

export async function GET(
  request: AuthenticatedRequest,
  { params }: { params: { id: string } }
) {
  // Authenticate request
  const authError = await authenticateRequest(request);
  if (authError) return authError;

  // Validate tenant
  const tenantError = await validateTenant(request);
  if (tenantError) return tenantError;

  try {
    const tenantId = request.headers.get('X-Tenant-ID')!;
    const emailId = params.id;

    const email = await prisma.email.findFirst({
      where: {
        id: emailId,
        tenant_id: tenantId,
      },
      include: {
        tasks: {
          select: {
            id: true,
            title: true,
            description: true,
            priority: true,
            status: true,
            created_at: true,
          },
        },
      },
    });

    if (!email) {
      return NextResponse.json(
        { error: 'Email não encontrado' },
        { status: 404 }
      );
    }

    return NextResponse.json(email);
  } catch (error) {
    console.error('Error fetching email:', error);
    return NextResponse.json(
      { error: 'Erro ao buscar email' },
      { status: 500 }
    );
  }
}
