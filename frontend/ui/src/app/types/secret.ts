import {BaseModel, UserAccount} from '@distr-sh/distr-sdk';

export interface Secret extends BaseModel {
  id: string;
  createdAt: string;
  updatedAt: string;
  updatedByUserAccountId?: string;
  customerOrganizationId?: string;
  key: string;
  updatedBy?: UserAccount;
}
