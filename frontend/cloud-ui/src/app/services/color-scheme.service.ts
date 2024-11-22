import {effect, Injectable} from '@angular/core';
import {BehaviorSubject, fromEvent} from 'rxjs';
import {toSignal} from '@angular/core/rxjs-interop';

@Injectable({
  providedIn: 'root',
})
export class ColorSchemeService {
  private _colorScheme: BehaviorSubject<'dark' | ''> = new BehaviorSubject<'dark' | ''>('');
  storage = toSignal(fromEvent(window, 'storage'));
  private COLOR_SCHEME = 'COLOR_SCHEME';

  constructor() {
    // syncs color scheme across tabs
    effect(() => {
      this.storage();
      this.initColorScheme();
    });
  }

  public initColorScheme() {
    switch (window.localStorage[this.COLOR_SCHEME]) {
      case '':
        this.updateColorScheme('');
        return;
      case 'dark':
        this.updateColorScheme('dark');
        return;
      default:
        if (window && window.matchMedia('(prefers-color-scheme: dark)').matches) this.updateColorScheme('dark');
    }
  }

  public colorScheme() {
    return this._colorScheme.asObservable();
  }

  public updateColorScheme(colorScheme: 'dark' | '') {
    window.localStorage[this.COLOR_SCHEME] = colorScheme;
    this._colorScheme.next(colorScheme);
  }
}
