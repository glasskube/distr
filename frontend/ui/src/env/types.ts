export interface Environment {
  production: boolean;
}

export interface RemoteEnvironment {
  readonly sentryDsn?: string;
  readonly sentryTraceSampleRate?: number;
  readonly posthogToken?: string;
  readonly posthogApiHost?: string;
  readonly posthogUiHost?: string;
  readonly registryHost: string;
}
