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
