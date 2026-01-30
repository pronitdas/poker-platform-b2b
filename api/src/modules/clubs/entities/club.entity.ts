import { Entity, Column, PrimaryGeneratedColumn, CreateDateColumn, UpdateDateColumn, ManyToOne, JoinColumn } from 'typeorm';
import { Agent } from '../../auth/entities/agent.entity';

@Entity('clubs')
export class Club {
  @PrimaryGeneratedColumn('uuid')
  id!: string;

  @Column()
  name!: string;

  @Column({ unique: true })
  code!: string;

  @Column('uuid')
  agentId!: string;

  @ManyToOne(() => Agent)
  @JoinColumn({ name: 'agent_id' })
  agent!: Agent;

  @Column({ type: 'decimal', precision: 5, scale: 2, default: 5 })
  rakePercent!: number;

  @Column({ default: true })
  isActive!: boolean;

  @Column({ type: 'jsonb', default: {} })
  settings!: Record<string, any>;

  @Column({ type: 'jsonb', default: {} })
  branding!: Record<string, any>;

  @CreateDateColumn({ name: 'created_at' })
  createdAt!: Date;

  @UpdateDateColumn({ name: 'updated_at' })
  updatedAt!: Date;
}
