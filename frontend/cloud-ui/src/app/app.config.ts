import {provideHttpClient, withInterceptors} from '@angular/common/http';
import {
  ApplicationConfig,
  ErrorHandler,
  inject,
  provideAppInitializer,
  provideZoneChangeDetection,
} from '@angular/core';
import {provideAnimationsAsync} from '@angular/platform-browser/animations/async';
import {provideRouter, Router} from '@angular/router';
import {routes} from './app.routes';
import {tokenInterceptor} from './services/auth.service';
import {errorToastInterceptor} from './services/error-toast.interceptor';
import {provideToastr} from 'ngx-toastr';
import * as Sentry from '@sentry/angular';

export const appConfig: ApplicationConfig = {
  providers: [
    {
      provide: ErrorHandler,
      useValue: Sentry.createErrorHandler(),
    },
    provideZoneChangeDetection({eventCoalescing: true}),
    provideRouter(routes),
    provideHttpClient(withInterceptors([tokenInterceptor, errorToastInterceptor])),
    provideAnimationsAsync(),
    provideToastr(),
  ],
};
