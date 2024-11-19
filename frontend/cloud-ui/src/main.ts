import { bootstrapApplication } from '@angular/platform-browser';
import { appConfig } from './app/app.config';
import { AppComponent } from './app/app.component';
import posthog from 'posthog-js';

posthog.init(
  '', // TODO: add api key
  {
    api_host:'https://eu.i.posthog.com',
    person_profiles: 'identified_only'
  }
)

bootstrapApplication(AppComponent, appConfig)
  .catch((err) => console.error(err));
