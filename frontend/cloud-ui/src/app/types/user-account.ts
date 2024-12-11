import {BaseModel} from './base';

export type UserRole = 'distributor' | 'customer';

export interface UserAccount extends BaseModel {
  email: string;
  name?: string;
}

export interface UserAccountWithRole extends UserAccount {
  userRole: UserRole;
}
