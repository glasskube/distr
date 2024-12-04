import {OverlayModule} from '@angular/cdk/overlay';
import {Component, inject, OnInit} from '@angular/core';
import {lastValueFrom} from 'rxjs';
import {dropdownAnimation} from '../../animations/dropdown';
import {AuthService} from '../../services/auth.service';
import {ColorSchemeSwitcherComponent} from '../color-scheme-switcher/color-scheme-switcher.component';

@Component({
  selector: 'app-nav-bar',
  standalone: true,
  templateUrl: './nav-bar.component.html',
  imports: [ColorSchemeSwitcherComponent, OverlayModule],
  animations: [dropdownAnimation],
})
export class NavBarComponent implements OnInit {
  private readonly auth = inject(AuthService);
  showDropdown = false;
  email?: string;
  name?: string;
  imageUrl?: string;

  public async ngOnInit() {
    try {
      const {email, name} = this.auth.getClaims();
      this.email = email;
      this.name = name;
      this.imageUrl = `https://www.gravatar.com/avatar/${await digestMessage(email)}`;
    } catch (e) {
      console.error(e);
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
