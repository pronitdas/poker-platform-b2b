import { Controller, Get, Post, Delete, Body, Param, UseGuards } from '@nestjs/common';
import { TablesService, TableConfig } from './tables.service';
import { JwtAuthGuard } from '../auth/jwt-auth.guard';

@Controller('tables')
export class TablesController {
  constructor(private tablesService: TablesService) {}

  @UseGuards(JwtAuthGuard)
  @Get()
  async findAll() {
    return this.tablesService.findAll();
  }

  @UseGuards(JwtAuthGuard)
  @Get(':tableId')
  async findOne(@Param('tableId') tableId: string) {
    return this.tablesService.findOne(tableId);
  }

  @UseGuards(JwtAuthGuard)
  @Get(':tableId/ws')
  async getGameServerUrl(@Param('tableId') tableId: string) {
    const url = await this.tablesService.getGameServerUrl(tableId);
    return { websocketUrl: url };
  }

  @UseGuards(JwtAuthGuard)
  @Post()
  async create(@Body() body: TableConfig) {
    return this.tablesService.create(body);
  }

  @UseGuards(JwtAuthGuard)
  @Delete(':tableId')
  async delete(@Param('tableId') tableId: string) {
    return this.tablesService.delete(tableId);
  }
}
