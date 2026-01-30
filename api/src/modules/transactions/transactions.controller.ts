import { Controller, Get, Post, Body, Query, UseGuards } from '@nestjs/common';
import { TransactionsService } from './transactions.service';
import { TransactionType } from './entities/transaction.entity';
import { JwtAuthGuard } from '../auth/jwt-auth.guard';

@Controller('transactions')
export class TransactionsController {
  constructor(private transactionsService: TransactionsService) {}

  @UseGuards(JwtAuthGuard)
  @Get()
  async findAll(
    @Query('playerId') playerId?: string,
    @Query('clubId') clubId?: string,
    @Query('type') type?: TransactionType,
  ) {
    return this.transactionsService.findAll({ playerId, clubId, type });
  }

  @UseGuards(JwtAuthGuard)
  @Post()
  async create(
    @Body() body: {
      playerId: string;
      clubId?: string;
      type: TransactionType;
      amount: number;
      reference?: string;
      metadata?: Record<string, any>;
    },
  ) {
    return this.transactionsService.create(body);
  }
}
