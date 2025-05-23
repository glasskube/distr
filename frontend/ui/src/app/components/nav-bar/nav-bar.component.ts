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
  faPlus,
  faShuffle,
  faStarOfLife,
} from '@fortawesome/free-solid-svg-icons';
import {UserRole} from '@glasskube/distr-sdk';
import {
  combineLatestWith,
  distinctUntilChanged,
  filter,
  firstValueFrom,
  lastValueFrom,
  map,
  Observable,
  Subject,
  takeUntil,
  withLatestFrom,
} from 'rxjs';
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
import {ContextService} from '../../services/context.service';

type SwitchOptions = {
  currentOrg: Organization;
  availableOrgs: OrganizationWithUserRole[];
  canCreateOrg: boolean;
};

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
  private destroyed$ = new Subject<void>();
  private readonly auth = inject(AuthService);
  public readonly sidebar = inject(SidebarService);
  private readonly toast = inject(ToastService);
  private readonly route = inject(ActivatedRoute);
  private readonly usersService = inject(UsersService);
  private readonly organizationService = inject(OrganizationService);
  private readonly organizationBranding = inject(OrganizationBrandingService);
  protected readonly organization$ = this.organizationService.get();
  protected readonly user$ = this.usersService.get();
  protected readonly switchOptions$: Observable<SwitchOptions> = this.organizationService.getAll().pipe(
    takeUntil(this.destroyed$),
    combineLatestWith(this.organization$),
    map(([orgs, currentOrg]) => {
      return {
        currentOrg,
        availableOrgs: orgs.filter((o) => o.id !== currentOrg.id),
        canCreateOrg: orgs.some((o) => o.userRole === 'vendor'),
      };
    })
  );

  userOpened = false;
  organizationsOpened = false;
  logoUrl = '/distr-logo.svg';
  customerSubtitle = 'Customer Portal';

  protected readonly faBarsStaggered = faBarsStaggered;
  protected tutorial?: string;

  public ngOnInit() {
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
      this.initBranding();
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
      const msg = getFormDisplayedError(e);
      if (msg) {
        this.toast.error(msg);
      }
    }
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
  protected readonly faPlus = faPlus;
}
