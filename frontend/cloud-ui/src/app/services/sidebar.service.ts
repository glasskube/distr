import {Injectable, signal} from '@angular/core';

@Injectable({providedIn: 'root'})
export class SidebarService {
  private showSidebarInternal = signal(false);
  public showSidebar = this.showSidebarInternal.asReadonly();

  public toggle(): void {
    this.showSidebarInternal.set(!this.showSidebarInternal());
  }

  public show(): void {
    this.showSidebarInternal.set(true);
  }

  public hide(): void {
    this.showSidebarInternal.set(false);
  }
}
