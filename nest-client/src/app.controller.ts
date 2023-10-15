import { Controller, Get, Query } from '@nestjs/common';
import { AppService } from './app.service';
import { MyData } from './app.model';

@Controller()
export class AppController {
  constructor(private readonly appService: AppService) {}

  @Get()
  async getHello(@Query() { limit }) {
    const result: MyData[] = [];

    for (let i = 0; i < Number(limit); i++) {
      const data: MyData = await this.appService.getData(i);
      result.push(data);
    }
    return result;
  }

  @Get('pong')
  async getSomething(@Query() { limit }) {
    const data: Promise<MyData>[] = [];
    for (let i = 0; i < Number(limit); i++) {
      data.push(this.appService.getData(i));
    }
    const response: MyData[] = await Promise.all(data);
    return response;
  }
}
