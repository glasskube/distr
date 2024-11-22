import {Component, EventEmitter, Input, Output} from '@angular/core';
import {ColorSchemeSwitcherComponent} from '../color-scheme-switcher/color-scheme-switcher.component';

@Component({
  selector: 'app-nav-bar',
  standalone: true,
  templateUrl: './nav-bar.component.html',
  imports: [ColorSchemeSwitcherComponent],
})
export class NavBarComponent {}
