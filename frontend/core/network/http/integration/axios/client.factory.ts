import axios from 'axios';

import type { IClientConfig } from '../../contract';
import { Client } from './client';

export class ClientFactory {
  static create(config?: IClientConfig) {
    const instance = axios.create(config);

    return new Client(instance);
  }
}
