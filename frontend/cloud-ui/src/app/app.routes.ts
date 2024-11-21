import {Routes} from '@angular/router';
import {DashboardPlaceholderComponent} from './components/dashboard-placeholder/dashboard-placeholder.component';
import {ApplicationsComponent} from './applications/applications.component';

export const routes: Routes = [
  {
    path: '',
    children: [
      {path: '', pathMatch: 'full', redirectTo: 'dashboard'},
      {path: 'dashboard', component: DashboardPlaceholderComponent},
      {path: 'applications', component: ApplicationsComponent},
    ],
  },
];
