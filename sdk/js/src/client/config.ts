export const defaultClientConfig = {apiBase: 'https://app.glasskube.cloud/api/v1/'};

export type ConditionalPartial<T, K extends keyof T> = Omit<T, K> & Partial<Pick<T, K>>;
