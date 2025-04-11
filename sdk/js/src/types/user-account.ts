import {BaseModel, Image} from './base';

export type UserRole = 'vendor' | 'customer';

export interface UserAccount extends BaseModel, Image {
  email: string;
  name?: string;
}

export interface UserAccountWithRole extends UserAccount {
  userRole: UserRole;
}
