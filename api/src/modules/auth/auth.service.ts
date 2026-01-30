import { Injectable, UnauthorizedException, ConflictException } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { JwtService } from '@nestjs/jwt';
import * as crypto from 'crypto';
import { Agent } from './entities/agent.entity';

@Injectable()
export class AuthService {
  constructor(
    @InjectRepository(Agent)
    private agentsRepository: Repository<Agent>,
    private jwtService: JwtService,
  ) {}

  private hashPassword(password: string, salt: string): string {
    return crypto.pbkdf2Sync(password, salt, 100000, 64, 'sha512').toString('hex');
  }

  private generateSalt(): string {
    return crypto.randomBytes(16).toString('hex');
  }

  async register(data: { email: string; password: string; companyName: string; contactName?: string }) {
    const existing = await this.agentsRepository.findOne({
      where: { email: data.email },
    });

    if (existing) {
      throw new ConflictException('Email already registered');
    }

    const salt = this.generateSalt();
    const passwordHash = this.hashPassword(data.password, salt);

    const agent = this.agentsRepository.create({
      email: data.email,
      passwordHash: `${salt}:${passwordHash}`,
      companyName: data.companyName,
      contactName: data.contactName,
    });

    await this.agentsRepository.save(agent);
    return { message: 'Registration successful', agentId: agent.id };
  }

  async login(email: string, password: string) {
    const agent = await this.agentsRepository.findOne({
      where: { email },
    });

    if (!agent) {
      throw new UnauthorizedException('Invalid credentials');
    }

    if (agent.status !== 'active') {
      throw new UnauthorizedException('Account is not active');
    }

    const [salt, hash] = agent.passwordHash.split(':');
    const passwordHash = this.hashPassword(password, salt);

    if (passwordHash !== hash) {
      throw new UnauthorizedException('Invalid credentials');
    }

    const payload = { sub: agent.id, email: agent.email, role: 'agent' };
    const token = this.jwtService.sign(payload);

    return {
      access_token: token,
      agent: {
        id: agent.id,
        email: agent.email,
        company_name: agent.companyName,
      },
    };
  }

  async validateAgent(agentId: string): Promise<Agent | null> {
    return this.agentsRepository.findOne({ where: { id: agentId } });
  }
}
