import {BaseModel, Named} from './base';

export interface Application extends BaseModel, Named {
  type: string;
  versions?: ApplicationVersion[];
}

export interface ApplicationVersion {
  id?: string;
  name?: string;
  createdAt?: string;
  applicationId?: string;
}
