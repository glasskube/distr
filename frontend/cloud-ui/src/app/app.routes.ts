import {Routes} from '@angular/router';
import {DashboardPlaceholderComponent} from './components/dashboard-placeholder/dashboard-placeholder.component';
import {ApplicationsPageComponent} from './applications/applications-page.component';
import {DeploymentTargetsPageComponent} from './deployment-targets/deployment-targets-page.component';

export const routes: Routes = [
  {
    path: '',
    children: [
      {path: '', pathMatch: 'full', redirectTo: 'dashboard'},
      {path: 'dashboard', component: DashboardPlaceholderComponent},
      {path: 'applications', component: ApplicationsPageComponent},
      {path: 'deployment-targets', component: DeploymentTargetsPageComponent},
    ],
  },
];
