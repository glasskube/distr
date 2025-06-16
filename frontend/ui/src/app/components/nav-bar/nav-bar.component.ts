import {OverlayModule} from '@angular/cdk/overlay';
import {HttpErrorResponse} from '@angular/common/http';
import {Component, inject, OnDestroy, OnInit, TemplateRef, ViewChild} from '@angular/core';
import {ActivatedRoute, RouterLink} from '@angular/router';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {
  faArrowLeft,
  faBarsStaggered,
  faCheck,
  faCheckDouble,
  faChevronDown,
  faChevronUp,
  faCircleExclamation,
  faClipboard,
  faLightbulb,
  faPlus,
  faShuffle,
  faXmark,
} from '@fortawesome/free-solid-svg-icons';
import {
  catchError,
  combineLatestWith,
  distinctUntilChanged,
  EMPTY,
  lastValueFrom,
  map,
  Observable,
  of,
  Subject,
  takeUntil,
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
import {OrganizationService} from '../../services/organization.service';
import {Organization, OrganizationWithUserRole} from '../../types/organization';
import {AutotrimDirective} from '../../directives/autotrim.directive';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {DialogRef, OverlayService} from '../../services/overlay.service';
import {modalFlyInOut} from '../../animations/modal';
import {UserAccount, UserAccountWithRole} from '../../../../../../sdk/js/src';

type SwitchOptions = {
  currentOrg: Organization;
  availableOrgs: OrganizationWithUserRole[];
  isVendorSomewhere: boolean;
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
    AutotrimDirective,
    ReactiveFormsModule,
  ],
  animations: [dropdownAnimation, modalFlyInOut],
})
export class NavBarComponent implements OnInit, OnDestroy {
  private destroyed$ = new Subject<void>();
  private readonly auth = inject(AuthService);
  private readonly overlay = inject(OverlayService);
  public readonly sidebar = inject(SidebarService);
  private readonly toast = inject(ToastService);
  private readonly route = inject(ActivatedRoute);
  private readonly usersService = inject(UsersService);
  private readonly organizationService = inject(OrganizationService);
  private readonly organizationBranding = inject(OrganizationBrandingService);
  protected readonly organization$ = this.organizationService.get();
  protected readonly user$ = this.usersService.get().pipe(
    catchError(() => {
      const claims = this.auth.getClaims();
      if (claims) {
        return of({
          id: claims.sub,
          name: claims.name,
          email: claims.email,
          userRole: claims.role,
          imageUrl: claims.image_url,
        });
      }
      return EMPTY;
    })
  );
  protected readonly switchOptions$: Observable<SwitchOptions> = this.organizationService.getAll().pipe(
    takeUntil(this.destroyed$),
    combineLatestWith(this.organization$),
    map(([orgs, currentOrg]) => {
      return {
        currentOrg,
        availableOrgs: orgs.filter((o) => o.id !== currentOrg.id),
        isVendorSomewhere: orgs.some((o) => o.userRole === 'vendor'),
      };
    })
  );

  userOpened = false;
  organizationsOpened = false;
  logoUrl = '/distr-logo.svg';
  customerSubtitle = 'Customer Portal';

  protected readonly faBarsStaggered = faBarsStaggered;
  protected tutorial?: string;

  @ViewChild('createOrgModal') private createOrgModal!: TemplateRef<unknown>;
  private modalRef?: DialogRef;
  protected readonly createOrgForm = new FormGroup({
    name: new FormControl<string>('', Validators.required),
  });

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

  async switchContext(org: Organization, targetPath = '/') {
    this.organizationsOpened = false;
    try {
      const switched = await lastValueFrom(this.auth.switchContext(org));
      if (switched) {
        location.assign(targetPath);
      }
    } catch (e) {
      const msg = getFormDisplayedError(e);
      if (msg) {
        this.toast.error(msg);
      }
    }
  }

  showCreateOrgModal(): void {
    this.closeCreateOrgModal();
    this.modalRef = this.overlay.showModal(this.createOrgModal);
    this.modalRef.addOnClosedHook((_) => {
      this.organizationsOpened = false;
    });
  }

  closeCreateOrgModal() {
    this.modalRef?.close();
    this.createOrgForm.reset();
  }

  async submitCreateOrgForm() {
    this.createOrgForm.markAllAsTouched();
    if (this.createOrgForm.valid) {
      try {
        const created = await lastValueFrom(
          this.organizationService.create({
            name: this.createOrgForm.value.name!,
            features: [],
          })
        );
        await this.switchContext(created, '/dashboard?from=new-org');
      } catch (e) {
        const msg = getFormDisplayedError(e);
        if (msg) {
          this.toast.error(msg);
        }
      }
    }
  }

  async logout() {
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
  protected readonly faCircleExclamation = faCircleExclamation;
  protected readonly faXmark = faXmark;
  protected readonly faClipboard = faClipboard;
}
