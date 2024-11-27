import {Component, ElementRef, ViewChild} from '@angular/core';
import {ApplicationsService} from './applications.service';
import {AsyncPipe, DatePipe} from '@angular/common';
import {Application} from '../types/application';
import {Observable} from 'rxjs';
import {FormArray, FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {
  faBoxArchive,
  faCaretDown,
  faMagnifyingGlass,
  faPen,
  faPlus,
  faTrash,
  faXmark,
} from '@fortawesome/free-solid-svg-icons';

@Component({
  selector: 'app-applications',
  standalone: true,
  imports: [AsyncPipe, DatePipe, ReactiveFormsModule, FaIconComponent],
  templateUrl: './applications.component.html',
})
export class ApplicationsComponent {
  magnifyingGlassIcon = faMagnifyingGlass;
  plusIcon = faPlus;
  caretDownIcon = faCaretDown;
  penIcon = faPen;
  trashIcon = faTrash;
  xmarkIcon = faXmark;
  releaseIcon = faBoxArchive;

  applications$!: Observable<Application[]>;
  selectedApplication?: Application;
  editForm = new FormGroup({
    id: new FormControl(''),
    name: new FormControl('', Validators.required),
    type: new FormControl('docker', Validators.required),
  });
  newVersionForm = new FormGroup({
    versionName: new FormControl('', Validators.required),
  });
  fileToUpload: File | null = null;
  @ViewChild('fileInput')
  fileInput?: ElementRef;

  public constructor(private applicationsService: ApplicationsService) {}

  ngOnInit() {
    this.applications$ = this.applicationsService.getApplications();
  }

  editApplication(application: Application) {
    this.selectedApplication = application;
    this.editForm.patchValue({
      id: application.id,
      name: application.name,
    });
    this.resetVersionForm();
  }

  newApplication() {
    this.selectedApplication = undefined;
    this.resetEditForm();
    this.resetVersionForm();
  }

  private resetEditForm() {
    this.editForm.reset();
    this.editForm.patchValue({type: 'docker'});
  }

  private resetVersionForm() {
    this.newVersionForm.reset();
    this.fileToUpload = null;
    if (this.fileInput) {
      this.fileInput.nativeElement.value = '';
    }
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

  onFileSelected(event: any) {
    this.fileToUpload = event.target.files[0];
  }

  createVersion() {
    if (this.newVersionForm.valid && this.fileToUpload != null && this.selectedApplication) {
      this.applicationsService
        .createApplicationVersion(
          this.selectedApplication,
          {
            name: this.newVersionForm.controls.versionName.value!,
          },
          this.fileToUpload
        )
        .subscribe((av) => {
          // not super correct state management, but good enough
          this.selectedApplication!.versions = [av, ...(this.selectedApplication!.versions || [])];
          this.resetVersionForm();
        });
    }
  }
}
