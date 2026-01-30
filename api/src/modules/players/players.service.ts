import { Injectable, NotFoundException } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { Player } from './entities/player.entity';

@Injectable()
export class PlayersService {
  constructor(
    @InjectRepository(Player)
    private playersRepository: Repository<Player>,
  ) {}

  async findAll(clubId?: string): Promise<Player[]> {
    const query = this.playersRepository.createQueryBuilder('player');
    if (clubId) {
      query.where('player.club_id = :clubId', { clubId });
    }
    return query.orderBy('player.created_at', 'DESC').getMany();
  }

  async findOne(id: string): Promise<Player> {
    const player = await this.playersRepository.findOne({ where: { id } });
    if (!player) {
      throw new NotFoundException(`Player ${id} not found`);
    }
    return player;
  }

  async create(data: { name: string; email: string; clubId?: string; balance?: number }): Promise<Player> {
    const player = this.playersRepository.create({
      name: data.name,
      email: data.email,
      clubId: data.clubId,
      balance: data.balance || 0,
    });
    return this.playersRepository.save(player);
  }

  async update(id: string, data: Partial<Player>): Promise<Player> {
    const player = await this.findOne(id);
    Object.assign(player, data);
    return this.playersRepository.save(player);
  }

  async delete(id: string): Promise<void> {
    const player = await this.findOne(id);
    await this.playersRepository.remove(player);
  }

  async updateBalance(id: string, amount: number): Promise<Player> {
    const player = await this.findOne(id);
    player.balance = Number(player.balance) + amount;
    return this.playersRepository.save(player);
  }
}
