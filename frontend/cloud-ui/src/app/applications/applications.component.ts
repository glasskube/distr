import {Component} from '@angular/core';
import {ApplicationsService} from './applications.service';
import {AsyncPipe, DatePipe, NgOptimizedImage} from '@angular/common';
import {Application} from '../types/application';
import {Observable} from 'rxjs';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faCaretDown, faMagnifyingGlass, faPen, faPlus, faTrash, faXmark} from '@fortawesome/free-solid-svg-icons';

@Component({
  selector: 'app-applications',
  standalone: true,
  imports: [AsyncPipe, DatePipe, ReactiveFormsModule, FaIconComponent, NgOptimizedImage],
  templateUrl: './applications.component.html',
})
export class ApplicationsComponent {
  magnifyingGlassIcon = faMagnifyingGlass;
  plusIcon = faPlus;
  caretDownIcon = faCaretDown;
  penIcon = faPen;
  trashIcon = faTrash;
  xmarkIcon = faXmark;

  applications$!: Observable<Application[]>;
  editForm = new FormGroup({
    id: new FormControl(''),
    name: new FormControl('', Validators.required),
    type: new FormControl('', Validators.required),
  });

  public constructor(private applicationsService: ApplicationsService) {}

  ngOnInit() {
    this.applications$ = this.applicationsService.getApplications();
  }

  editApplication(application: Application) {
    this.editForm.patchValue({
      id: application.id,
      name: application.name,
      type: 'docker',
    });
  }

  newApplication() {
    this.editForm.reset();
  }

  saveApplication() {
    if (this.editForm.valid) {
      const val = this.editForm.value;
      let result: Observable<Application>;
      if (!val.id) {
        result = this.applicationsService.createApplication({
          name: val.name!,
          type: val.type!,
        });
      } else {
        result = this.applicationsService.updateApplication({
          id: val.id!,
          name: val.name!,
        });
      }
      result.subscribe((application) => this.editApplication(application));
    }
  }
}
