import { Injectable, NotFoundException } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { Club } from './entities/club.entity';

@Injectable()
export class ClubsService {
  constructor(
    @InjectRepository(Club)
    private clubsRepository: Repository<Club>,
  ) {}

  async findAll(agentId?: string): Promise<Club[]> {
    const query = this.clubsRepository.createQueryBuilder('club');
    if (agentId) {
      query.where('club.agent_id = :agentId', { agentId });
    }
    return query.orderBy('club.created_at', 'DESC').getMany();
  }

  async findOne(id: string): Promise<Club> {
    const club = await this.clubsRepository.findOne({ where: { id } });
    if (!club) {
      throw new NotFoundException(`Club ${id} not found`);
    }
    return club;
  }

  async create(data: { name: string; code: string; agentId: string; rakePercent?: number }): Promise<Club> {
    const club = this.clubsRepository.create({
      name: data.name,
      code: data.code,
      agentId: data.agentId,
      rakePercent: data.rakePercent || 5,
    });
    return this.clubsRepository.save(club);
  }

  async update(id: string, data: Partial<Club>): Promise<Club> {
    const club = await this.findOne(id);
    Object.assign(club, data);
    return this.clubsRepository.save(club);
  }

  async delete(id: string): Promise<void> {
    const club = await this.findOne(id);
    await this.clubsRepository.remove(club);
  }
}
