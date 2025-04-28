import {OverlayModule} from '@angular/cdk/overlay';
import {HttpErrorResponse} from '@angular/common/http';
import {Component, inject, OnInit} from '@angular/core';
import {RouterLink} from '@angular/router';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faBarsStaggered} from '@fortawesome/free-solid-svg-icons';
import {UserRole} from '@glasskube/distr-sdk';
import {lastValueFrom} from 'rxjs';
import {digestMessage} from '../../../util/crypto';
import {getFormDisplayedError} from '../../../util/errors';
import {dropdownAnimation} from '../../animations/dropdown';
import {AuthService} from '../../services/auth.service';
import {OrganizationBrandingService} from '../../services/organization-branding.service';
import {SidebarService} from '../../services/sidebar.service';
import {ToastService} from '../../services/toast.service';
import {ColorSchemeSwitcherComponent} from '../color-scheme-switcher/color-scheme-switcher.component';

@Component({
  selector: 'app-nav-bar',
  standalone: true,
  templateUrl: './nav-bar.component.html',
  imports: [ColorSchemeSwitcherComponent, OverlayModule, FaIconComponent, RouterLink],
  animations: [dropdownAnimation],
})
export class NavBarComponent implements OnInit {
  private readonly auth = inject(AuthService);
  public readonly sidebar = inject(SidebarService);
  private readonly toast = inject(ToastService);
  private readonly organizationBranding = inject(OrganizationBrandingService);
  showDropdown = false;
  email?: string;
  name?: string;
  role?: UserRole;
  imageUrl?: string;
  logoUrl = '/distr-logo.svg';
  customerSubtitle = 'Customer Portal';

  protected readonly faBarsStaggered = faBarsStaggered;

  public async ngOnInit() {
    try {
      const claims = this.auth.getClaims();
      if (claims) {
        const {email, name, role} = claims;
        this.email = email;
        this.name = name;
        this.role = role;
        this.initBranding();
        this.imageUrl = `https://www.gravatar.com/avatar/${await digestMessage(email)}`;
      }
    } catch (e) {
      console.error(e);
    }
  }

  private async initBranding() {
    if (this.auth.hasRole('customer')) {
      try {
        const branding = await lastValueFrom(this.organizationBranding.get());
        if (branding.logo) {
          this.logoUrl = `data:${branding.logoContentType};base64,${branding.logo}`;
        }
        if (branding.title) {
          this.customerSubtitle = branding.title;
        }
      } catch (e) {
        const msg = getFormDisplayedError(e);
        if (msg && e instanceof HttpErrorResponse && e.status !== 404) {
          this.toast.error(msg);
        }
      }
    }
  }

  public async logout() {
    await lastValueFrom(this.auth.logout());
    // This is necessary to flush the caching crud services
    // TODO: implement flushing of services directly and switch to router.navigate(...)
    location.assign('/login');
  }
}
