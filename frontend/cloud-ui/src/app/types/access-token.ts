import {BaseModel} from './base';

export interface AccessToken extends BaseModel {
  expiresAt?: string;
  lastUsedAt?: string;
  label?: string;
}

export interface AccessTokenWithKey extends AccessToken {
  key: string;
}

export interface CreateAccessTokenRequest {
  label?: string;
  expiresAt?: Date;
}
