import {AsyncPipe, DatePipe, NgOptimizedImage} from '@angular/common';
import {Component, ElementRef, inject, Input, TemplateRef, ViewChild, ViewContainerRef} from '@angular/core';
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
import {OverlayModule} from '@angular/cdk/overlay';
import {dropdownAnimation} from '../animations/dropdown';
import {EmbeddedOverlayRef, OverlayService} from '../services/overlay.service';
import {drawerFlyInOut} from '../animations/drawer';

@Component({
  selector: 'app-applications',
  imports: [AsyncPipe, DatePipe, ReactiveFormsModule, FaIconComponent, NgOptimizedImage, OverlayModule],
  templateUrl: './applications.component.html',
  animations: [dropdownAnimation, drawerFlyInOut],
})
export class ApplicationsComponent {
  @Input('fullVersion') fullVersion: boolean = false;
  magnifyingGlassIcon = faMagnifyingGlass;
  plusIcon = faPlus;
  caretDownIcon = faCaretDown;
  penIcon = faPen;
  trashIcon = faTrash;
  xmarkIcon = faXmark;
  releaseIcon = faBoxArchive;
  showDropdown = false;

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

  private manageApplicationDrawerRef?: EmbeddedOverlayRef;
  private readonly overlay = inject(OverlayService);
  private readonly viewContainerRef = inject(ViewContainerRef);

  openDrawer(templateRef: TemplateRef<unknown>, application?: Application) {
    this.hideDrawer();
    if (application) {
      this.loadApplication(application);
    } else {
      this.reset();
    }
    this.manageApplicationDrawerRef = this.overlay.showDrawer(templateRef, this.viewContainerRef);
  }

  hideDrawer() {
    this.manageApplicationDrawerRef?.close();
    this.reset();
  }

  loadApplication(application: Application) {
    this.selectedApplication = application;
    this.editForm.patchValue({
      id: application.id,
      name: application.name,
    });
    this.resetVersionForm();
  }

  reset() {
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
      result.subscribe((application) => this.loadApplication(application));
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
