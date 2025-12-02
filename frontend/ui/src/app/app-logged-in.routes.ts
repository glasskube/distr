import {inject} from '@angular/core';
import {CanActivateFn, Router, Routes} from '@angular/router';
import dayjs from 'dayjs';
import {firstValueFrom, map} from 'rxjs';
import {UserRole} from '../../../../sdk/js/src';
import {getRemoteEnvironment} from '../env/remote';
import {AccessTokensComponent} from './access-tokens/access-tokens.component';
import {ApplicationDetailComponent} from './applications/application-detail.component';
import {ApplicationsPageComponent} from './applications/applications-page.component';
import {ArtifactLicensesComponent} from './artifacts/artifact-licenses/artifact-licenses.component';
import {ArtifactPullsComponent} from './artifacts/artifact-pulls/artifact-pulls.component';
import {ArtifactVersionsComponent} from './artifacts/artifact-versions/artifact-versions.component';
import {ArtifactsComponent} from './artifacts/artifacts/artifacts.component';
import {CustomerOrganizationsComponent} from './components/customer-organizations/customer-organizations.component';
import {DashboardComponent} from './components/dashboard/dashboard.component';
import {HomeComponent} from './components/home/home.component';
import {CustomerUsersComponent} from './components/users/customers/customer-users.component';
import {VendorUsersComponent} from './components/users/vendors/vendor-users.component';
import {DeploymentTargetsComponent} from './deployments/deployment-targets.component';
import {LicensesComponent} from './licenses/licenses.component';
import {OrganizationBrandingComponent} from './organization-branding/organization-branding.component';
import {OrganizationSettingsComponent} from './organization-settings/organization-settings.component';
import {AuthService} from './services/auth.service';
import {FeatureFlagService} from './services/feature-flag.service';
import {OrganizationService} from './services/organization.service';
import {ToastService} from './services/toast.service';
import {SubscriptionCallbackComponent} from './subscription/subscription-callback.component';
import {SubscriptionComponent} from './subscription/subscription.component';
import {AgentsTutorialComponent} from './tutorials/agents/agents-tutorial.component';
import {BrandingTutorialComponent} from './tutorials/branding/branding-tutorial.component';
import {RegistryTutorialComponent} from './tutorials/registry/registry-tutorial.component';
import {TutorialsComponent} from './tutorials/tutorials.component';

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

function registryHostSetOrRedirectGuard(redirectTo: string): CanActivateFn {
  return async () => {
    const router = inject(Router);
    const toast = inject(ToastService);
    const env = await getRemoteEnvironment();
    if ((env.registryHost ?? '').length > 0) {
      return true;
    }
    toast.error('Registry must be enabled first!');
    return router.createUrlTree([redirectTo]);
  };
}

function subscriptionGuard(): CanActivateFn {
  return () => {
    const auth = inject(AuthService);
    const router = inject(Router);
    const organizationService = inject(OrganizationService);
    return (
      auth.hasRole('customer') ||
      organizationService
        .get()
        .pipe(map((org) => (dayjs(org.subscriptionEndsAt).isBefore() ? router.createUrlTree(['/subscription']) : true)))
    );
  };
}

export const routes: Routes = [
  {
    path: '',
    canActivate: [subscriptionGuard()],
    children: [
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
      {path: 'deployments', component: DeploymentTargetsComponent},
      {
        path: 'artifacts',
        children: [
          {path: '', pathMatch: 'full', component: ArtifactsComponent},
          {path: ':id', component: ArtifactVersionsComponent},
        ],
      },
      {
        path: 'artifact-pulls',
        component: ArtifactPullsComponent,
        canActivate: [requiredRoleGuard('vendor')],
      },
      {
        path: 'customers',
        component: CustomerOrganizationsComponent,
        canActivate: [requiredRoleGuard('vendor')],
      },
      {
        path: 'customers/:customerOrganizationId',
        component: CustomerUsersComponent,
        canActivate: [requiredRoleGuard('vendor')],
      },
      {
        path: 'users',
        component: VendorUsersComponent,
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
        canActivate: [requiredRoleGuard('vendor'), licensingEnabledGuard()],
        data: {userRole: 'vendor'},
        children: [
          {
            path: 'applications',
            component: LicensesComponent,
          },
          {
            path: 'artifacts',
            component: ArtifactLicensesComponent,
          },
        ],
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
      {
        path: 'tutorials',
        canActivate: [requiredRoleGuard('vendor')],
        children: [
          {
            path: '',
            pathMatch: 'full',
            component: TutorialsComponent,
          },
          {
            path: 'agents',
            component: AgentsTutorialComponent,
          },
          {
            path: 'branding',
            component: BrandingTutorialComponent,
          },
          {
            path: 'registry',
            canActivate: [registryHostSetOrRedirectGuard('/tutorials')],
            component: RegistryTutorialComponent,
          },
        ],
      },
    ],
  },
  {
    path: 'subscription',
    canActivate: [requiredRoleGuard('vendor')],
    children: [
      {
        path: '',
        pathMatch: 'full',
        component: SubscriptionComponent,
      },
      {
        path: 'callback',
        component: SubscriptionCallbackComponent,
      },
    ],
  },
];
