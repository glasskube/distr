import {provideHttpClient, withInterceptors} from '@angular/common/http';
import {ApplicationConfig, inject, provideZoneChangeDetection} from '@angular/core';
import {provideAnimationsAsync} from '@angular/platform-browser/animations/async';
import {provideRouter} from '@angular/router';
import {routes} from './app.routes';
import {AuthService, tokenInterceptor} from './services/auth.service';
import {errorInterceptor} from './services/error.interceptor';
import {provideToastr} from 'ngx-toastr';
import {ToastComponent} from './components/toast.component';

export const appConfig: ApplicationConfig = {
  providers: [
    provideZoneChangeDetection({eventCoalescing: true}),
    provideRouter(routes),
    provideHttpClient(withInterceptors([tokenInterceptor, errorInterceptor])),
    provideAnimationsAsync(),
    provideToastr(),
  ],
};
