export const defaultClientConfig = {apiBase: 'https://app.distr.sh/api/v1/'};

export type ConditionalPartial<T, K extends keyof T> = Omit<T, K> & Partial<Pick<T, K>>;
