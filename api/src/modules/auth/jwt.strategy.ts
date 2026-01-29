import { Injectable, UnauthorizedException } from '@nestjs/common';
import { PassportStrategy } from '@nestjs/passport';
import { ExtractJwt, Strategy } from 'passport-jwt';
import { AuthService } from './auth.service';

@Injectable()
export class JwtStrategy extends PassportStrategy(Strategy) {
  constructor(private authService: AuthService) {
    super({
      jwtFromRequest: ExtractJwt.fromAuthHeaderAsBearerToken(),
      ignoreExpiration: false,
      secretOrKey: process.env.JWT_SECRET || 'your-super-secret-key-change-in-production',
    });
  }

  async validate(payload: { sub: string; email: string; role: string }) {
    const agent = await this.authService.validateAgent(payload.sub);
    if (!agent) {
      throw new UnauthorizedException();
    }
    return { id: agent.id, email: agent.email, role: payload.role };
  }
}
