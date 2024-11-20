import {bootstrapApplication} from '@angular/platform-browser';
import {appConfig} from './app/app.config';
import {AppComponent} from './app/app.component';
import posthog from 'posthog-js';

posthog.init('phc_tQTyjXLct9rmpLrFKo7HDLBDXERBnfviHpyzeWL9wTy', {
  api_host: 'https://p.glasskube.eu',
  ui_host: 'https://eu.i.posthog.com',
  person_profiles: 'identified_only',
  // pageview event capturing is done for Angular router events.
  // Here we prevent the window "load" event from triggering a duplicate pageview event.
  capture_pageview: false,
});

bootstrapApplication(AppComponent, appConfig).catch((err) => console.error(err));
