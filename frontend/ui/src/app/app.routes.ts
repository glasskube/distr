import {inject} from '@angular/core';
import {
  ActivatedRouteSnapshot,
  CanActivateFn,
  createUrlTreeFromSnapshot,
  Router,
  RouterStateSnapshot,
  Routes,
} from '@angular/router';
import {firstValueFrom} from 'rxjs';
import {ApplicationsPageComponent} from './applications/applications-page.component';
import {NavShellComponent} from './components/nav-shell.component';
import {UsersComponent} from './components/users/users.component';
import {DeploymentsPageComponent} from './deployments/deployments-page.component';
import {ForgotComponent} from './forgot/forgot.component';
import {InviteComponent} from './invite/invite.component';
import {LoginComponent} from './login/login.component';
import {PasswordResetComponent} from './password-reset/password-reset.component';
import {RegisterComponent} from './register/register.component';
import {AuthService} from './services/auth.service';
import {SettingsService} from './services/settings.service';
import {ToastService} from './services/toast.service';
import {VerifyComponent} from './verify/verify.component';
import {OrganizationBrandingComponent} from './organization-branding/organization-branding.component';
import {UserRole} from '@glasskube/distr-sdk';
import {AccessTokensComponent} from './access-tokens/access-tokens.component';
import {LicensesComponent} from './licenses/licenses.component';
import {FeatureFlagService} from './services/feature-flag.service';
import {ApplicationDetailComponent} from './applications/application-detail.component';
import {UsersService} from './services/users.service';

const emailVerificationGuard: CanActivateFn = async () => {
  const auth = inject(AuthService);
  const settings = inject(SettingsService);
  const toast = inject(ToastService);
  const router = inject(Router);
  const claims = auth.getClaims();
  if (claims?.email_verified) {
    await firstValueFrom(settings.confirmEmailVerification());
    toast.success('Your email has been verified');
    await firstValueFrom(auth.logout());
    return router.createUrlTree(['/login'], {queryParams: {email: claims.email}});
  }
  return true;
};

const jwtParamRedirectGuard: CanActivateFn = (route: ActivatedRouteSnapshot) => {
  const auth = inject(AuthService);
  const jwt = route.queryParamMap.get('jwt');
  if (jwt === null) {
    return true;
  } else {
    // TODO: flush crud service caches
    auth.actionToken = jwt;
    const newtree = createUrlTreeFromSnapshot(route, [], null, null);
    delete newtree.queryParams['jwt']; // prevent infinite loop
    return newtree;
  }
};

const jwtAuthGuard: CanActivateFn = (route: ActivatedRouteSnapshot, state: RouterStateSnapshot) => {
  const auth = inject(AuthService);
  const router = inject(Router);
  const claims = auth.getClaims();
  if (claims) {
    if (claims.password_reset) {
      if (state.url === '/reset') {
        return true;
      } else {
        return router.createUrlTree(['/reset']);
      }
    } else if (!claims.email_verified) {
      if (state.url === '/verify') {
        return true;
      } else {
        return router.createUrlTree(['/verify']);
      }
    } else {
      return true;
    }
  } else {
    return router.createUrlTree(['/login']);
  }
};

const inviteComponentGuard: CanActivateFn = async () => {
  const auth = inject(AuthService);
  const users = inject(UsersService);
  const router = inject(Router);
  try {
    const {active} = await firstValueFrom(users.getUserStatus());
    if (!active) {
      return true;
    }
  } catch (e) {}
  auth.actionToken = null;
  return router.createUrlTree(['/login']);
};

function requiredRoleGuard(userRole: UserRole): CanActivateFn {
  return () => inject(AuthService).hasRole(userRole);
}

function licensingEnabledGuard(): CanActivateFn {
  return async () => {
    const featureFlags = inject(FeatureFlagService);
    return await firstValueFrom(featureFlags.isLicensingEnabled$);
  };
}

const baseRouteRedirectGuard: CanActivateFn = () => {
  const auth = inject(AuthService);
  const router = inject(Router);
  switch (auth.getClaims()?.role) {
    case 'customer':
      return router.createUrlTree(['/home']);
    case 'vendor':
      return router.createUrlTree(['/dashboard']);
    default:
      return false;
  }
};

export const routes: Routes = [
  {path: 'login', component: LoginComponent},
  {path: 'register', component: RegisterComponent},
  {path: 'forgot', component: ForgotComponent},
  {
    path: '',
    canActivate: [jwtParamRedirectGuard, jwtAuthGuard],
    children: [
      {
        path: '',
        pathMatch: 'full',
        canActivate: [baseRouteRedirectGuard],
        children: [],
      },
      {
        path: 'verify',
        component: VerifyComponent,
        canActivate: [emailVerificationGuard],
      },
      {path: 'reset', component: PasswordResetComponent},
      {path: 'join', component: InviteComponent, canActivate: [inviteComponentGuard]},
      {
        path: '',
        component: NavShellComponent,
        children: [
          {
            path: 'dashboard',
            loadComponent: async () =>
              (await import('./components/dashboard-placeholder/dashboard-placeholder.component'))
                .DashboardPlaceholderComponent,
            canActivate: [requiredRoleGuard('vendor')],
          },
          {
            path: 'home',
            loadComponent: async () => (await import('./components/home/home.component')).HomeComponent,
            canActivate: [requiredRoleGuard('customer')],
          },
          {
            path: 'applications',
            canActivate: [requiredRoleGuard('vendor')],
            children: [
              {
                path: '',
                pathMatch: 'full',
                component: ApplicationsPageComponent,
              },
              {
                path: ':applicationId',
                component: ApplicationDetailComponent,
              },
            ],
          },
          {path: 'deployments', component: DeploymentsPageComponent},
          {
            path: 'customers',
            component: UsersComponent,
            data: {userRole: 'customer'},
            canActivate: [requiredRoleGuard('vendor')],
          },
          {
            path: 'users',
            component: UsersComponent,
            data: {userRole: 'vendor'},
            canActivate: [requiredRoleGuard('vendor')],
          },
          {
            path: 'branding',
            component: OrganizationBrandingComponent,
            data: {userRole: 'vendor'},
            canActivate: [requiredRoleGuard('vendor')],
          },
          {
            path: 'licenses',
            component: LicensesComponent,
            data: {userRole: 'vendor'},
            canActivate: [requiredRoleGuard('vendor'), licensingEnabledGuard()],
          },
          {
            path: 'settings',
            children: [
              {
                path: 'access-tokens',
                component: AccessTokensComponent,
              },
            ],
          },
        ],
      },
    ],
  },
];
