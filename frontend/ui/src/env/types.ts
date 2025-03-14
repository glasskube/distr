export interface Environment {
  production: boolean;
}

export interface RemoteEnvironment {
  readonly sentryDsn?: string;
  readonly posthogToken?: string;
  readonly posthogApiHost?: string;
  readonly posthogUiHost?: string;
  readonly artifactsHost: string;
}
