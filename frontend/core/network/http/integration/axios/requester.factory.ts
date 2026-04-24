import axios from 'axios';

import type { IRequesterConfig } from '../../contract';
import { Requester } from './requester';

export class RequesterFactory {
  static create(config?: IRequesterConfig) {
    const instance = axios.create(config);

    return new Requester(instance);
  }
}
