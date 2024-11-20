import {Component, OnInit} from '@angular/core';
import {SideBarComponent} from './components/side-bar/side-bar.component';
import {initFlowbite} from 'flowbite';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [SideBarComponent],
  templateUrl: './app.component.html',
  styleUrl: './app.component.scss'
})
export class AppComponent implements OnInit {
  title = 'Glasskube Cloud';

  ngOnInit(): void {
    initFlowbite();
  }

}
