import {BaseModel, Named} from './base';

export interface Application extends BaseModel, Named {
  type?: string;
}
