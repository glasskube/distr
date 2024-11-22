import {Component, inject} from '@angular/core';
import {NgIf} from '@angular/common';
import {ColorSchemeService} from '../../services/color-scheme.service';

@Component({
  selector: 'app-color-scheme-switcher',
  standalone: true,
  templateUrl: './color-scheme-switcher.component.html',
  imports: [NgIf],
})
export class ColorSchemeSwitcherComponent {
  private colorSchemeService = inject(ColorSchemeService);
  public colorSchemeSignal = this.colorSchemeService.colorScheme;

  constructor() {}

  switchColorScheme() {
    let newColorScheme: 'dark' | '' = 'dark';
    if ('dark' === this.colorSchemeSignal()) {
      newColorScheme = '';
    }
    this.colorSchemeSignal.set(newColorScheme);
  }
}
