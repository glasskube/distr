import {bootstrapApplication} from '@angular/platform-browser';
import dayjs from 'dayjs';
import duration from 'dayjs/plugin/duration';
import relativeTime from 'dayjs/plugin/relativeTime';
import posthog from 'posthog-js';
import {AppComponent} from './app/app.component';
import {appConfig} from './app/app.config';
import {environment} from './env/env';

dayjs.extend(duration);
dayjs.extend(relativeTime);

if (environment.posthogToken) {
  posthog.init(environment.posthogToken, {
    api_host: 'https://p.glasskube.eu',
    ui_host: 'https://eu.i.posthog.com',
    person_profiles: 'identified_only',
    // pageview event capturing is done for Angular router events.
    // Here we prevent the window "load" event from triggering a duplicate pageview event.
    capture_pageview: false,
  });
}

bootstrapApplication(AppComponent, appConfig).catch((err) => console.error(err));
