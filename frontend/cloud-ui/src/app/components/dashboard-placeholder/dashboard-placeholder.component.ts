import {Component} from '@angular/core';
import {initFlowbite} from 'flowbite';

@Component({
  selector: 'app-dashboard-placeholder',
  standalone: true,
  templateUrl: './dashboard-placeholder.component.html',
})
export class DashboardPlaceholderComponent {
  ngAfterViewInit() {
    initFlowbite();
  }
}
