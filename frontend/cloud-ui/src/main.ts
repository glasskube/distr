import {bootstrapApplication} from '@angular/platform-browser';
import dayjs from 'dayjs';
import duration from 'dayjs/plugin/duration';
import relativeTime from 'dayjs/plugin/relativeTime';
import posthog from 'posthog-js';
import {AppComponent} from './app/app.component';
import {appConfig} from './app/app.config';
import {environment} from './env/env';
import * as Sentry from '@sentry/angular';
import {buildConfig} from './buildconfig';

Sentry.init({
  enabled: environment.production,
  release: buildConfig.version ?? buildConfig.commit,
  dsn: 'https://2a42d7067e57e6d98bf5bec1737c6020@o4508443344633856.ingest.de.sentry.io/4508443366719568',
  integrations: [],
});

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
    before_send: [
      (cr) => {
        if (cr !== null) {
          if (cr.$set === undefined) {
            cr.$set = {};
          }
          if (cr.$set_once === undefined) {
            cr.$set_once = {};
          }
          cr.$set['version'] = buildConfig.version;
          cr.$set_once['version'] = buildConfig.version;
        }
        return cr;
      },
    ],
  });
}

bootstrapApplication(AppComponent, appConfig).catch((err) => console.error(err));
