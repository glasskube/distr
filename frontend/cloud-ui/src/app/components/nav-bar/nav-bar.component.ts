import {OverlayModule} from '@angular/cdk/overlay';
import {Component, inject, OnInit} from '@angular/core';
import {lastValueFrom} from 'rxjs';
import {dropdownAnimation} from '../../animations/dropdown';
import {AuthService} from '../../services/auth.service';
import {SidebarService} from '../../services/sidebar.service';
import {ColorSchemeSwitcherComponent} from '../color-scheme-switcher/color-scheme-switcher.component';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faBarsStaggered} from '@fortawesome/free-solid-svg-icons';
import {UserRole} from '../../types/user-account';
import {RouterLink} from '@angular/router';
import {OrganizationBrandingService} from '../../services/organization-branding.service';

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
  private readonly organizationBranding = inject(OrganizationBrandingService);
  showDropdown = false;
  email?: string;
  name?: string;
  role?: UserRole;
  imageUrl?: string;
  logoUrl = '/glasskube-logo.svg';
  customerSubtitle = 'Customer Portal';

  protected readonly faBarsStaggered = faBarsStaggered;

  public async ngOnInit() {
    try {
      const {email, name, role} = this.auth.getClaims();
      this.email = email;
      this.name = name;
      this.role = role;
      this.initBranding();
      this.imageUrl = `https://www.gravatar.com/avatar/${await digestMessage(email)}`;
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
        console.error(e);
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

/**
 * From https://developer.mozilla.org/en-US/docs/Web/API/SubtleCrypto/digest
 */
async function digestMessage(message: string, alg: string = 'SHA-256'): Promise<string> {
  const msgUint8 = new TextEncoder().encode(message); // encode as (utf-8) Uint8Array
  const hashBuffer = await crypto.subtle.digest(alg, msgUint8); // hash the message
  const hashArray = Array.from(new Uint8Array(hashBuffer)); // convert buffer to byte array
  const hashHex = hashArray.map((b) => b.toString(16).padStart(2, '0')).join(''); // convert bytes to hex string
  return hashHex;
}
