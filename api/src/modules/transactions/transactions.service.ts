import { Injectable } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { Transaction, TransactionType, TransactionStatus } from './entities/transaction.entity';

@Injectable()
export class TransactionsService {
  constructor(
    @InjectRepository(Transaction)
    private transactionsRepository: Repository<Transaction>,
  ) {}

  async findAll(filters?: { playerId?: string; clubId?: string; type?: TransactionType }): Promise<Transaction[]> {
    const query = this.transactionsRepository.createQueryBuilder('tx');
    if (filters?.playerId) {
      query.where('tx.player_id = :playerId', { playerId: filters.playerId });
    }
    if (filters?.clubId) {
      query.andWhere('tx.club_id = :clubId', { clubId: filters.clubId });
    }
    if (filters?.type) {
      query.andWhere('tx.type = :type', { type: filters.type });
    }
    return query.orderBy('tx.created_at', 'DESC').getMany();
  }

  async create(data: {
    playerId: string;
    clubId?: string;
    type: TransactionType;
    amount: number;
    reference?: string;
    metadata?: Record<string, any>;
  }): Promise<Transaction> {
    const transaction = this.transactionsRepository.create({
      ...data,
      status: TransactionStatus.COMPLETED,
    });
    return this.transactionsRepository.save(transaction);
  }
}
