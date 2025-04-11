import { NextRequest, NextResponse } from 'next/server';
import { authenticateRequest, validateTenant, AuthenticatedRequest } from '../middleware';
import { prisma } from '@/lib/prisma';
import config from '@/lib/config';

export async function GET(request: AuthenticatedRequest) {
  // Authenticate request
  const authError = await authenticateRequest(request);
  if (authError) return authError;

  // Validate tenant
  const tenantError = await validateTenant(request);
  if (tenantError) return tenantError;

  try {
    const searchParams = request.nextUrl.searchParams;
    const tenantId = request.headers.get('X-Tenant-ID')!;

    const where = {
      tenant_id: tenantId,
      ...(searchParams.get('category') && { category: searchParams.get('category') }),
      ...(searchParams.get('priority') && { priority: searchParams.get('priority') }),
      ...(searchParams.get('start_date') && {
        created_at: { gte: new Date(searchParams.get('start_date')!) },
      }),
      ...(searchParams.get('end_date') && {
        created_at: { lte: new Date(searchParams.get('end_date')!) },
      }),
    };

    const emails = await prisma.email.findMany({
      where,
      orderBy: { created_at: 'desc' },
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

    return NextResponse.json(emails);
  } catch (error) {
    console.error('Error fetching emails:', error);
    return NextResponse.json(
      { error: 'Erro ao buscar emails' },
      { status: 500 }
    );
  }
}

export async function POST(request: AuthenticatedRequest) {
  // Authenticate request
  const authError = await authenticateRequest(request);
  if (authError) return authError;

  // Validate tenant
  const tenantError = await validateTenant(request);
  if (tenantError) return tenantError;

  try {
    const emailData = await request.json();
    const tenantId = request.headers.get('X-Tenant-ID')!;

    // Call Go service for classification
    const classificationResponse = await fetch(`${config.api.classifierUrl}/api/v1/classify`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
      },
      body: JSON.stringify({
        subject: emailData.subject,
        body: emailData.body,
        sender: emailData.sender,
        recipient: emailData.recipient,
      }),
    });

    if (!classificationResponse.ok) {
      throw new Error('Erro ao classificar email');
    }

    const classification = await classificationResponse.json();

    // Save email with classification and tasks in a transaction
    const email = await prisma.$transaction(async (tx: typeof prisma) => {
      // Create email
      const email = await tx.email.create({
        data: {
          tenant: {
            connect: {
              id: tenantId
            }
          },
          subject: emailData.subject,
          body: emailData.body,
          sender: emailData.sender,
          recipient: emailData.recipient,
          priority: classification.priority,
          category: classification.category,
          labels: classification.labels || [],
        },
      });

      // Create tasks if any were suggested
      if (classification.suggestedTasks?.length > 0) {
        for (const task of classification.suggestedTasks) {
          await tx.task.create({
            data: {
              email: {
                connect: { id: email.id }
              },
              tenant: {
                connect: { id: tenantId }
              },
              title: task.title,
              description: task.description,
              priority: task.priority,
              status: 'pending',
            },
          });
        }
      }

      // Return email with created tasks
      return tx.email.findUnique({
        where: { id: email.id },
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
    });

    return NextResponse.json(email);
  } catch (error) {
    console.error('Error processing email:', error);
    return NextResponse.json(
      { error: 'Erro ao processar email' },
      { status: 500 }
    );
  }
}
