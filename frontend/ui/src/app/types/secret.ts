import {BaseModel, UserAccount} from '@glasskube/distr-sdk';

export interface Secret extends BaseModel {
  id: string;
  createdAt: string;
  updatedAt: string;
  updatedByUserAccountId?: string;
  customerOrganizationId?: string;
  key: string;
  updatedBy?: UserAccount;
}
