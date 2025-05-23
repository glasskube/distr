import {BaseModel} from './base';

export type UserRole = 'vendor' | 'customer';

export interface UserAccount extends BaseModel {
  email: string;
  name?: string;
  imageUrl?: string;
}

export interface UserAccountWithRole extends UserAccount {
  userRole: UserRole;
  joinedOrgAt: string;
}
