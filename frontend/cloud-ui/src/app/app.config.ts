import {provideHttpClient, withInterceptors} from '@angular/common/http';
import {ApplicationConfig, provideZoneChangeDetection} from '@angular/core';
import {provideAnimationsAsync} from '@angular/platform-browser/animations/async';
import {provideRouter} from '@angular/router';
import {routes} from './app.routes';
import {tokenInterceptor} from './services/auth.service';
import {errorToastInterceptor} from './services/error-toast.interceptor';
import {provideToastr} from 'ngx-toastr';

export const appConfig: ApplicationConfig = {
  providers: [
    provideZoneChangeDetection({eventCoalescing: true}),
    provideRouter(routes),
    provideHttpClient(withInterceptors([tokenInterceptor, errorToastInterceptor])),
    provideAnimationsAsync(),
    provideToastr(),
  ],
};
