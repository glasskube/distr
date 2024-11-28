import {OverlayModule} from '@angular/cdk/overlay';
import {Component} from '@angular/core';
import {dropdownAnimation} from '../../animations/dropdown';
import {ColorSchemeSwitcherComponent} from '../color-scheme-switcher/color-scheme-switcher.component';

@Component({
  selector: 'app-nav-bar',
  standalone: true,
  templateUrl: './nav-bar.component.html',
  imports: [ColorSchemeSwitcherComponent, OverlayModule],
  animations: [dropdownAnimation],
})
export class NavBarComponent {
  showDropdown = false;
}
