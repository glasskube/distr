import {BaseModel, Named} from './base';
import {Geolocation} from './geolocation';

export interface DeploymentTarget extends BaseModel, Named {
  name: string;
  type: string;
  geolocation?: Geolocation;
}
