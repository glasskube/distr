import {BaseModel} from './base';

export interface CustomerOrganization extends Required<BaseModel> {
  name: string;
  imageId?: string;
  imageUrl?: string;
}

export interface CustomerOrganizationWithUsage extends CustomerOrganization {
  userCount: number;
  deploymentTargetCount: number;
}

export interface CreateUpdateCustomerOrganizationRequest {
  name: string;
  imageId?: string;
}
