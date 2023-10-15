import { Injectable } from '@nestjs/common';
import axios from 'axios';
import { MyData } from './app.model';

@Injectable()
export class AppService {
  async getData(id: number) {
    try {
      const { data } = await axios.get<Promise<MyData>>(
        `http://localhost:8080/api/v1/get_something?id=${id}`,
        { timeout: 10000 },
      );
      return data;
    } catch (error) {
      console.error(error);
      return null;
    }
  }
}
