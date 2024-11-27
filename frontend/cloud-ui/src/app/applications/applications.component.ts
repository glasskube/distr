import {AsyncPipe, DatePipe, NgOptimizedImage} from '@angular/common';
import {Component, ElementRef, inject, ViewChild} from '@angular/core';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
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
import {Observable} from 'rxjs';
import {Application} from '../types/application';
import {ApplicationsService} from '../services/applications.service';

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
  releaseIcon = faBoxArchive;

  private readonly applications = inject(ApplicationsService);
  applications$: Observable<Application[]> = this.applications.list();
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
        result = this.applications.create({
          name: val.name!,
          type: val.type!,
        });
      } else {
        result = this.applications.update({
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
      this.applications
        .createApplicationVersion(
          this.selectedApplication,
          {
            name: this.newVersionForm.controls.versionName.value!,
          },
          this.fileToUpload
        )
        .subscribe((av) => {
          this.resetVersionForm();
        });
    }
  }
}
