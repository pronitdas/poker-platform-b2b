import { Injectable } from '@nestjs/common';

export interface DashboardStats {
  totalAgents: number;
  totalClubs: number;
  totalPlayers: number;
  totalTables: number;
  activeGamesToday: number;
  revenueToday: number;
}

@Injectable()
export class AdminService {
  async getDashboardStats(): Promise<DashboardStats> {
    // In production, this would query the database
    return {
      totalAgents: 5,
      totalClubs: 12,
      totalPlayers: 1500,
      totalTables: 45,
      activeGamesToday: 320,
      revenueToday: 15000.50,
    };
  }

  async getSystemHealth(): Promise<{ status: string; components: Record<string, string> }> {
    return {
      status: 'healthy',
      components: {
        api: 'up',
        gameServer: 'up',
        database: 'up',
        cache: 'up',
        kafka: 'up',
      },
    };
  }
}
