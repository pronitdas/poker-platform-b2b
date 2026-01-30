import { Injectable } from '@nestjs/common';

export interface TableConfig {
  tableId: string;
  gameType: string;
  bettingType: string;
  maxPlayers: number;
  minPlayers: number;
  smallBlind: number;
  bigBlind: number;
  buyInMin: number;
  buyInMax: number;
}

@Injectable()
export class TablesService {
  private tables: Map<string, TableConfig> = new Map();

  async findAll(): Promise<TableConfig[]> {
    return Array.from(this.tables.values());
  }

  async findOne(tableId: string): Promise<TableConfig | null> {
    return this.tables.get(tableId) || null;
  }

  async create(data: TableConfig): Promise<TableConfig> {
    this.tables.set(data.tableId, data);
    return data;
  }

  async delete(tableId: string): Promise<void> {
    this.tables.delete(tableId);
  }

  async getGameServerUrl(tableId: string): Promise<string | null> {
    const table = await this.findOne(tableId);
    if (!table) {
      return null;
    }
    return `ws://localhost:3002/ws/${tableId}`;
  }
}
