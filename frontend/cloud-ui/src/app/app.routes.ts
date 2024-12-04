import {Router, Routes} from '@angular/router';
import {DashboardPlaceholderComponent} from './components/dashboard-placeholder/dashboard-placeholder.component';
import {ApplicationsPageComponent} from './applications/applications-page.component';
import {DeploymentTargetsPageComponent} from './deployment-targets/deployment-targets-page.component';
import {inject} from '@angular/core';
import {AuthService} from './services/auth.service';
import {LoginComponent} from './login/login.component';
import {NavShellComponent} from './components/nav-shell.component';

export const routes: Routes = [
  {path: 'login', component: LoginComponent},
  {
    path: '',
    canActivate: [
      () => {
        const auth = inject(AuthService);
        const router = inject(Router);
        if (auth.isAuthenticated) {
          return true;
        } else {
          return router.createUrlTree(['/login']);
        }
      },
    ],
    component: NavShellComponent,
    children: [
      {path: '', pathMatch: 'full', redirectTo: 'dashboard'},
      {path: 'dashboard', component: DashboardPlaceholderComponent},
      {path: 'applications', component: ApplicationsPageComponent},
      {path: 'deployment-targets', component: DeploymentTargetsPageComponent},
    ],
  },
];
