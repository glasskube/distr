import {Component} from '@angular/core';
import {SideBarComponent} from './components/side-bar/side-bar.component';
import {NavBarComponent} from './components/nav-bar/nav-bar.component';
import {RouterOutlet} from '@angular/router';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [SideBarComponent, NavBarComponent, RouterOutlet],
  templateUrl: './app.component.html',
  styleUrl: './app.component.scss'
})
export class AppComponent {
  title = 'Glasskube Cloud';
}
