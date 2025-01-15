import {BaseModel} from './base';

export interface OrganizationBranding extends BaseModel {
  title?: string;
  description?: string;
  logo?: string;
  logoFileName?: string;
  logoContentType?: string;
}
