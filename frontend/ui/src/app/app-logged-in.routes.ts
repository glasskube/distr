import {DashboardComponent} from './components/dashboard/dashboard.component';
import {HomeComponent} from './components/home/home.component';
import {ApplicationsPageComponent} from './applications/applications-page.component';
import {ApplicationDetailComponent} from './applications/application-detail.component';
import {DeploymentsPageComponent} from './deployments/deployments-page.component';
import {ArtifactsComponent} from './artifacts/artifacts/artifacts.component';
import {ArtifactVersionsComponent} from './artifacts/artifact-versions/artifact-versions.component';
import {ArtifactLicensesComponent} from './artifacts/artifact-licenses/artifact-licenses.component';
import {ArtifactPullsComponent} from './artifacts/artifact-pulls/artifact-pulls.component';
import {UsersComponent} from './components/users/users.component';
import {OrganizationBrandingComponent} from './organization-branding/organization-branding.component';
import {OrganizationSettingsComponent} from './organization-settings/organization-settings.component';
import {LicensesComponent} from './licenses/licenses.component';
import {AccessTokensComponent} from './access-tokens/access-tokens.component';
import {CanActivateFn, Router, Routes} from '@angular/router';
import {UserRole} from '../../../../sdk/js/src';
import {inject} from '@angular/core';
import {AuthService} from './services/auth.service';
import {FeatureFlagService} from './services/feature-flag.service';
import {firstValueFrom} from 'rxjs';

function requiredRoleGuard(userRole: UserRole): CanActivateFn {
  return () => {
    if (inject(AuthService).hasRole(userRole)) {
      return true;
    }
    return inject(Router).createUrlTree(['/']);
  };
}

function licensingEnabledGuard(): CanActivateFn {
  return async () => {
    const featureFlags = inject(FeatureFlagService);
    return await firstValueFrom(featureFlags.isLicensingEnabled$);
  };
}

function registryEnabledGuard(): CanActivateFn {
  return async () => {
    const featureFlags = inject(FeatureFlagService);
    return await firstValueFrom(featureFlags.isRegistryEnabled$);
  };
}

export const routes: Routes = [
  {
    path: 'dashboard',
    component: DashboardComponent,
    canActivate: [requiredRoleGuard('vendor')],
  },
  {
    path: 'home',
    component: HomeComponent,
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
    path: 'artifacts',
    children: [
      {path: '', pathMatch: 'full', component: ArtifactsComponent},
      {path: ':id', component: ArtifactVersionsComponent},
    ],
    canActivate: [registryEnabledGuard()],
  },
  {
    path: 'artifact-licenses',
    children: [{path: '', pathMatch: 'full', component: ArtifactLicensesComponent}],
    data: {userRole: 'vendor'},
    canActivate: [requiredRoleGuard('vendor'), registryEnabledGuard()],
  },
  {
    path: 'artifact-pulls',
    component: ArtifactPullsComponent,
    canActivate: [requiredRoleGuard('vendor'), registryEnabledGuard()],
  },
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
    path: 'settings',
    component: OrganizationSettingsComponent,
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
];
