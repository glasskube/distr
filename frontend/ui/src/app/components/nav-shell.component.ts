import {Component} from '@angular/core';
import {RouterOutlet} from '@angular/router';
import {NavBarComponent} from './nav-bar/nav-bar.component';
import {SideBarComponent} from './side-bar/side-bar.component';

@Component({
  selector: 'app-nav-shell',
  template: `
    <app-nav-bar></app-nav-bar>
    <app-side-bar></app-side-bar>
    <router-outlet></router-outlet>
  `,
  imports: [NavBarComponent, SideBarComponent, RouterOutlet],
})
export class NavShellComponent {}
