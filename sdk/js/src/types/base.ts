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

export interface Image {
  id?: string;
  image?: string;
  imageFileName?: string;
  imageContentType?: string;
}
