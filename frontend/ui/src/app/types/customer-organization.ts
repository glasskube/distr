import {BaseModel} from '@glasskube/distr-sdk';

export interface CustomerOrganization extends Required<BaseModel> {
  name: string;
  imageId?: string;
  imageUrl?: string;
}

export interface CustomerOrganizationWithUserCount extends CustomerOrganization {
  userCount: number;
}

export interface CreateUpdateCustomerOrganizationRequest {
  name: string;
  imageId?: string;
}
