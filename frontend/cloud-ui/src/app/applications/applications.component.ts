import {Component} from '@angular/core';
import { ApplicationsService } from './applications.service';
import { AsyncPipe } from '@angular/common';
import { Application } from '../types/application';
import { Observable } from 'rxjs';

@Component({
  selector: 'app-applications',
  standalone: true,
  imports: [AsyncPipe],
  templateUrl: './applications.component.html',
  styleUrl: './applications.component.scss',
})
export class ApplicationsComponent {
  applications$!: Observable<Application[]>;
  public constructor(private applicationsService: ApplicationsService) {}

  ngOnInit() {
    this.applications$ = this.applicationsService.getApplications();
  }
}
