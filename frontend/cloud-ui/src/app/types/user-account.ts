import {BaseModel} from './base';

export interface UserAccount extends BaseModel {
  email: string;
  name?: string;
}
