import {ActivatedRouteSnapshot, createUrlTreeFromSnapshot, Router, RouterStateSnapshot, Routes} from '@angular/router';
import {DashboardPlaceholderComponent} from './components/dashboard-placeholder/dashboard-placeholder.component';
import {ApplicationsPageComponent} from './applications/applications-page.component';
import {inject} from '@angular/core';
import {AuthService} from './services/auth.service';
import {LoginComponent} from './login/login.component';
import {NavShellComponent} from './components/nav-shell.component';
import {RegisterComponent} from './register/register.component';
import {InviteComponent} from './invite/invite.component';
import {UsersComponent} from './components/users/users.component';
import {DeploymentsPageComponent} from './deployments/deployments-page.component';
import {VerifyComponent} from './verify/verify.component';
import {ForgotComponent} from './forgot/forgot.component';
import {PasswordResetComponent} from './password-reset/password-reset.component';

export const routes: Routes = [
  {path: 'login', component: LoginComponent},
  {path: 'register', component: RegisterComponent},
  {path: 'forgot', component: ForgotComponent},
  {
    path: '',
    canActivate: [
      (route: ActivatedRouteSnapshot) => {
        const auth = inject(AuthService);
        const jwt = route.queryParamMap.get('jwt');
        if (jwt === null) {
          return true;
        } else {
          // TODO: flush crud service caches
          auth.token = jwt;
          const newtree = createUrlTreeFromSnapshot(route, [], null, null);
          delete newtree.queryParams['jwt']; // prevent infinite loop
          return newtree;
        }
      },
      (route: ActivatedRouteSnapshot, state: RouterStateSnapshot) => {
        const auth = inject(AuthService);
        const router = inject(Router);
        if (auth.isAuthenticated) {
          if (auth.getClaims().password_reset) {
            if (state.url === '/reset') {
              return true;
            } else {
              return router.createUrlTree(['/reset']);
            }
          } else if (!auth.getClaims().email_verified) {
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
      },
    ],
    children: [
      {path: '', pathMatch: 'full', redirectTo: 'dashboard'},
      {path: 'verify', component: VerifyComponent},
      {path: 'reset', component: PasswordResetComponent},
      {path: 'join', component: InviteComponent},
      {
        path: '',
        component: NavShellComponent,
        children: [
          {path: 'dashboard', component: DashboardPlaceholderComponent},
          {path: 'applications', component: ApplicationsPageComponent},
          {path: 'deployments', component: DeploymentsPageComponent},
          {path: 'customers', component: UsersComponent, data: {userRole: 'customer'}},
          {path: 'users', component: UsersComponent, data: {userRole: 'vendor'}},
        ],
      },
    ],
  },
];
