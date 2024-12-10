import {Component, effect, ElementRef, inject, ViewChild} from '@angular/core';
import {RouterLink} from '@angular/router';
import {SidebarService} from '../../services/sidebar.service';

@Component({
  selector: 'app-side-bar',
  standalone: true,
  templateUrl: './side-bar.component.html',
  imports: [RouterLink],
})
export class SideBarComponent {
  public readonly sidebar = inject(SidebarService);
  public feedbackAlert = true;

  @ViewChild('asideElement') private asideElement!: ElementRef<HTMLElement>;

  constructor() {
    effect(() => {
      const show = this.sidebar.showSidebar();
      this.asideElement.nativeElement.classList.toggle('translate-x-0', show);
      this.asideElement.nativeElement.classList.toggle('-translate-x-full', !show);
    });
  }
}
