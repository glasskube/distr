import {BaseModel} from './base';
import {UserRole} from './user-account';

export type AccessTokenSecretSlot = 1 | 2;

export interface AccessTokenSecret {
  slot: AccessTokenSecretSlot;
  createdAt: string;
  lastUsedAt?: string;
}

export interface AccessToken extends BaseModel {
  expiresAt?: string;
  lastUsedAt?: string;
  label?: string;
  userRole?: UserRole;
  secrets: AccessTokenSecret[];
}

export interface AccessTokenWithKey extends AccessToken {
  key: string;
}

export interface CreateAccessTokenRequest {
  label?: string;
  expiresAt?: Date;
  userRole?: UserRole;
}
