import {Component, inject} from '@angular/core';
import {NgIf} from '@angular/common';
import {ColorSchemeService} from '../../services/color-scheme.service';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faMoon, faSun} from '@fortawesome/free-solid-svg-icons';

@Component({
  selector: 'app-color-scheme-switcher',
  standalone: true,
  templateUrl: './color-scheme-switcher.component.html',
  imports: [NgIf, FaIconComponent],
})
export class ColorSchemeSwitcherComponent {
  private colorSchemeService = inject(ColorSchemeService);
  public colorSchemeSignal = this.colorSchemeService.colorScheme;

  protected readonly faSun = faSun;
  protected readonly faMoon = faMoon;

  constructor() {}

  switchColorScheme() {
    let newColorScheme: 'dark' | '' = 'dark';
    if ('dark' === this.colorSchemeSignal()) {
      newColorScheme = '';
    }
    this.colorSchemeSignal.set(newColorScheme);
  }
}
