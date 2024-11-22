import {Component, inject, OnInit} from '@angular/core';
import {AsyncPipe, NgIf} from '@angular/common';
import {ColorSchemeService} from '../../services/color-scheme.service';

@Component({
  selector: 'app-color-scheme-switcher',
  standalone: true,
  templateUrl: './color-scheme-switcher.component.html',
  imports: [NgIf, AsyncPipe],
})
export class ColorSchemeSwitcherComponent implements OnInit {
  private colorSchemeService = inject(ColorSchemeService);
  colorScheme$ = this.colorSchemeService.colorScheme();
  colorScheme: 'dark' | '' = '';

  switchColorScheme() {
    let newColorScheme: 'dark' | '' = 'dark';
    if ('dark' === this.colorScheme) {
      newColorScheme = '';
    }
    this.colorScheme = newColorScheme;
    this.colorSchemeService.updateColorScheme(newColorScheme);
  }

  ngOnInit(): void {
    this.colorSchemeService.colorScheme().subscribe((colorScheme) => (this.colorScheme = colorScheme));
    this.colorSchemeService.initColorScheme();
  }
}
