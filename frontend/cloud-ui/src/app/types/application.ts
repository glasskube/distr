export interface Application {
  id?: string;
  name?: string;
  createdAt?: string;
  type?: string;
  versions?: ApplicationVersion[];
}

export interface ApplicationVersion {
  id?: string;
  name?: string;
  createdAt?: string;
}
