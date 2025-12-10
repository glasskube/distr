import {effect, Injectable, signal, WritableSignal} from '@angular/core';
import {toSignal} from '@angular/core/rxjs-interop';
import {fromEvent} from 'rxjs';

@Injectable({
  providedIn: 'root',
})
export class ColorSchemeService {
  private COLOR_SCHEME = 'COLOR_SCHEME';

  // syncs color scheme across tabs
  private storageSignal = toSignal(fromEvent(window, 'storage'));
  private _colorSchemeSignal: WritableSignal<'dark' | ''> = signal(this.readColorSchemeFromLocalStorage());

  public get colorScheme() {
    return this._colorSchemeSignal;
  }

  constructor() {
    effect(() => {
      window.localStorage[this.COLOR_SCHEME] = this.colorScheme();
    });
    effect(() => {
      this.storageSignal();
      this.colorScheme.set(this.readColorSchemeFromLocalStorage());
    });
  }

  private readColorSchemeFromLocalStorage() {
    switch (window.localStorage[this.COLOR_SCHEME]) {
      case '':
        return '';
      case 'dark':
        return 'dark';
      default:
        if (window && window.matchMedia('(prefers-color-scheme: dark)').matches) {
          return 'dark';
        }
    }
    return '';
  }
}
