import {BaseModel, Named} from './base';

export interface DeploymentTarget extends BaseModel, Named {
  name: string;
  type: string;
  geolocation?: Geolocation;
}
