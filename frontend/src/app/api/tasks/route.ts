import { NextRequest, NextResponse } from 'next/server';
import { authenticateRequest, validateTenant, AuthenticatedRequest } from '../middleware';
import { prisma } from '@/lib/prisma';

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

    const page = parseInt(searchParams.get('page') || '1');
    const pageSize = Math.min(parseInt(searchParams.get('page_size') || '10'), 100);
    const skip = (page - 1) * pageSize;

    const where = {
      tenant_id: tenantId,
      ...(searchParams.get('email_id') && { email_id: searchParams.get('email_id') }),
      ...(searchParams.get('user_id') && { assigned_to: searchParams.get('user_id') }),
      ...(searchParams.get('priority') && { priority: searchParams.get('priority') }),
      ...(searchParams.get('status') && { status: searchParams.get('status') }),
      ...(searchParams.get('start_date') && {
        created_at: { gte: new Date(searchParams.get('start_date')!) },
      }),
      ...(searchParams.get('end_date') && {
        created_at: { lte: new Date(searchParams.get('end_date')!) },
      }),
    };

    const [tasks, total] = await Promise.all([
      prisma.task.findMany({
        where,
        skip,
        take: pageSize,
        orderBy: { created_at: 'desc' },
        include: {
          email: {
            select: {
              subject: true,
              sender: true,
            },
          },
        },
      }),
      prisma.task.count({ where }),
    ]);

    return NextResponse.json({
      data: tasks,
      pagination: {
        total,
        page,
        pageSize,
        totalPages: Math.ceil(total / pageSize),
      },
    });
  } catch (error) {
    console.error('Error fetching tasks:', error);
    return NextResponse.json(
      { error: 'Erro ao buscar tarefas' },
      { status: 500 }
    );
  }
}

export async function PUT(request: AuthenticatedRequest) {
  // Authenticate request
  const authError = await authenticateRequest(request);
  if (authError) return authError;

  // Validate tenant
  const tenantError = await validateTenant(request);
  if (tenantError) return tenantError;

  try {
    const tenantId = request.headers.get('X-Tenant-ID')!;
    const data = await request.json();
    const { id, ...updateData } = data;

    const task = await prisma.task.findFirst({
      where: {
        id,
        tenant_id: tenantId,
      },
    });

    if (!task) {
      return NextResponse.json(
        { error: 'Tarefa não encontrada' },
        { status: 404 }
      );
    }

    const updatedTask = await prisma.task.update({
      where: { id },
      data: {
        ...updateData,
        updated_at: new Date(),
      },
      include: {
        email: {
          select: {
            subject: true,
            sender: true,
          },
        },
      },
    });

    return NextResponse.json(updatedTask);
  } catch (error) {
    console.error('Error updating task:', error);
    return NextResponse.json(
      { error: 'Erro ao atualizar tarefa' },
      { status: 500 }
    );
  }
}

export async function DELETE(request: AuthenticatedRequest) {
  // Authenticate request
  const authError = await authenticateRequest(request);
  if (authError) return authError;

  // Validate tenant
  const tenantError = await validateTenant(request);
  if (tenantError) return tenantError;

  try {
    const tenantId = request.headers.get('X-Tenant-ID')!;
    const { id } = await request.json();

    const task = await prisma.task.findFirst({
      where: {
        id,
        tenant_id: tenantId,
      },
    });

    if (!task) {
      return NextResponse.json(
        { error: 'Tarefa não encontrada' },
        { status: 404 }
      );
    }

    await prisma.task.delete({
      where: { id },
    });

    return new NextResponse(null, { status: 204 });
  } catch (error) {
    console.error('Error deleting task:', error);
    return NextResponse.json(
      { error: 'Erro ao deletar tarefa' },
      { status: 500 }
    );
  }
}
