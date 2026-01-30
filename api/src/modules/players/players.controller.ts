import { Controller, Get, Post, Put, Delete, Body, Param, UseGuards } from '@nestjs/common';
import { PlayersService } from './players.service';
import { Player } from './entities/player.entity';
import { JwtAuthGuard } from '../auth/jwt-auth.guard';

@Controller('players')
export class PlayersController {
  constructor(private playersService: PlayersService) {}

  @UseGuards(JwtAuthGuard)
  @Get()
  async findAll() {
    return this.playersService.findAll();
  }

  @UseGuards(JwtAuthGuard)
  @Get(':id')
  async findOne(@Param('id') id: string) {
    return this.playersService.findOne(id);
  }

  @UseGuards(JwtAuthGuard)
  @Post()
  async create(@Body() body: { name: string; email: string; clubId?: string; balance?: number }) {
    return this.playersService.create(body);
  }

  @UseGuards(JwtAuthGuard)
  @Put(':id')
  async update(@Param('id') id: string, @Body() body: Partial<Player>) {
    return this.playersService.update(id, body);
  }

  @UseGuards(JwtAuthGuard)
  @Delete(':id')
  async delete(@Param('id') id: string) {
    return this.playersService.delete(id);
  }

  @UseGuards(JwtAuthGuard)
  @Put(':id/balance')
  async updateBalance(@Param('id') id: string, @Body() body: { amount: number }) {
    return this.playersService.updateBalance(id, body.amount);
  }
}
