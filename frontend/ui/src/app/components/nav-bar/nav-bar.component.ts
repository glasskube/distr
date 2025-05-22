import {OverlayModule} from '@angular/cdk/overlay';
import {HttpErrorResponse} from '@angular/common/http';
import {Component, inject, OnDestroy, OnInit} from '@angular/core';
import {ActivatedRoute, RouterLink} from '@angular/router';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {
  faArrowLeft,
  faBarsStaggered,
  faCheck,
  faCheckDouble,
  faChevronDown,
  faChevronUp,
  faLightbulb,
  faShuffle,
  faStarOfLife,
} from '@fortawesome/free-solid-svg-icons';
import {UserRole} from '@glasskube/distr-sdk';
import {distinctUntilChanged, filter, firstValueFrom, lastValueFrom, map, Subject, takeUntil} from 'rxjs';
import {getFormDisplayedError} from '../../../util/errors';
import {dropdownAnimation} from '../../animations/dropdown';
import {AuthService} from '../../services/auth.service';
import {OrganizationBrandingService} from '../../services/organization-branding.service';
import {SidebarService} from '../../services/sidebar.service';
import {ToastService} from '../../services/toast.service';
import {ColorSchemeSwitcherComponent} from '../color-scheme-switcher/color-scheme-switcher.component';
import {UsersService} from '../../services/users.service';
import {SecureImagePipe} from '../../../util/secureImage';
import {AsyncPipe, TitleCasePipe} from '@angular/common';
import {RequireRoleDirective} from '../../directives/required-role.directive';
import {OrganizationService} from '../../services/organization.service';
import {Organization, OrganizationWithUserRole} from '../../types/organization';
import {faCircleCheck} from '@fortawesome/free-regular-svg-icons';

@Component({
  selector: 'app-nav-bar',
  standalone: true,
  templateUrl: './nav-bar.component.html',
  imports: [
    ColorSchemeSwitcherComponent,
    OverlayModule,
    FaIconComponent,
    RouterLink,
    SecureImagePipe,
    AsyncPipe,
    TitleCasePipe,
  ],
  animations: [dropdownAnimation],
})
export class NavBarComponent implements OnInit, OnDestroy {
  private readonly auth = inject(AuthService);
  public readonly sidebar = inject(SidebarService);
  private readonly toast = inject(ToastService);
  private readonly users = inject(UsersService);
  private readonly route = inject(ActivatedRoute);
  private readonly organizationService = inject(OrganizationService);
  protected readonly organizations$ = this.organizationService.list();

  private readonly organizationBranding = inject(OrganizationBrandingService);
  userOpened = false;
  organizationsOpened = false;
  currentOrg = this.organizationService.get();
  email?: string;
  name?: string;
  role?: UserRole;
  imageUrl?: string;
  logoUrl = '/distr-logo.svg';
  customerSubtitle = 'Customer Portal';

  protected readonly faBarsStaggered = faBarsStaggered;
  private destroyed$ = new Subject<void>();
  protected tutorial?: string;

  public async ngOnInit() {
    this.route.queryParams
      .pipe(
        map((params) => params['tutorial']),
        distinctUntilChanged(),
        takeUntil(this.destroyed$)
      )
      .subscribe((tutorial) => {
        this.tutorial = tutorial;
      });

    try {
      const claims = this.auth.getClaims();
      if (claims) {
        const {email, name, role} = claims;
        this.email = email;
        this.name = name;
        this.role = role;
        this.initBranding();
        this.imageUrl = (await firstValueFrom(this.users.getUserByEmail(email))).imageUrl;
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

  async switchContext(org: OrganizationWithUserRole) {
    this.organizationsOpened = false;
    try {
      const switched = await lastValueFrom(this.auth.switchContext(org));
      if (switched) {
        location.assign('/');
      }
    } catch (e) {
      console.log(e);
      const msg = getFormDisplayedError(e);
      if (msg) {
        this.toast.error(msg);
      }
    }
  }

  isCurrentOrg(org: OrganizationWithUserRole): boolean {
    return org.id === this.auth.getClaims()?.org;
  }

  public async logout() {
    await lastValueFrom(this.auth.logout());
    // This is necessary to flush the caching crud services
    // TODO: implement flushing of services directly and switch to router.navigate(...)
    location.assign('/login');
  }

  ngOnDestroy() {
    this.destroyed$.next();
    this.destroyed$.complete();
  }

  protected readonly faLightbulb = faLightbulb;
  protected readonly faArrowLeft = faArrowLeft;
  protected readonly faShuffle = faShuffle;
  protected readonly faCheck = faCheck;
  protected readonly faCheckDouble = faCheckDouble;
  protected readonly faChevronDown = faChevronDown;
  protected readonly faChevronUp = faChevronUp;
}
