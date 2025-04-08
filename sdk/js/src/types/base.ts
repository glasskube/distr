export interface BaseModel {
  id?: string;
  createdAt?: string;
}

export interface Named {
  name?: string;
}

export interface TokenResponse {
  token: string;
}

export interface DeploymentTargetAccessResponse {
  connectUrl: string;
  targetId: string;
  targetSecret: string;
}

export interface WithIcon {
  id: string;
  icon?: string;
  iconFileName?: string;
  iconContentType?: string;
}
