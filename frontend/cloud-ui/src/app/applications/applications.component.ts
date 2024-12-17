import {OverlayModule} from '@angular/cdk/overlay';
import {AsyncPipe, DatePipe, NgOptimizedImage} from '@angular/common';
import {Component, ElementRef, inject, Input, OnDestroy, OnInit, TemplateRef, ViewChild} from '@angular/core';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faBoxArchive, faMagnifyingGlass, faPen, faPlus, faTrash, faXmark} from '@fortawesome/free-solid-svg-icons';
import {EMPTY, filter, map, Observable, startWith, Subject, switchMap, takeUntil, withLatestFrom} from 'rxjs';
import {drawerFlyInOut} from '../animations/drawer';
import {dropdownAnimation} from '../animations/dropdown';
import {modalFlyInOut} from '../animations/modal';
import {RequireRoleDirective} from '../directives/required-role.directive';
import {ApplicationsService} from '../services/applications.service';
import {DialogRef, OverlayService} from '../services/overlay.service';
import {ToastService} from '../services/toast.service';
import {Application} from '../types/application';
import {combineLatest} from 'rxjs';
import {filteredByFormControl} from '../../util/filter';

@Component({
  selector: 'app-applications',
  imports: [
    AsyncPipe,
    DatePipe,
    ReactiveFormsModule,
    FaIconComponent,
    NgOptimizedImage,
    OverlayModule,
    RequireRoleDirective,
  ],
  templateUrl: './applications.component.html',
  animations: [dropdownAnimation, drawerFlyInOut, modalFlyInOut],
})
export class ApplicationsComponent implements OnInit, OnDestroy {
  @Input('fullVersion') fullVersion: boolean = false;
  protected readonly faMagnifyingGlass = faMagnifyingGlass;
  protected readonly faPlus = faPlus;
  protected readonly faPen = faPen;
  protected readonly faXmark = faXmark;
  protected readonly faBoxArchive = faBoxArchive;
  protected readonly faTrash = faTrash;
  showDropdown = false;

  private readonly destroyed$ = new Subject<void>();
  private readonly applications = inject(ApplicationsService);
  filterForm = new FormGroup({
    search: new FormControl(''),
  });
  applications$: Observable<Application[]> = filteredByFormControl(
    this.applications.list(),
    this.filterForm.controls.search,
    (it: Application, search: string) => !search || it.name?.toLowerCase().indexOf(search.toLowerCase()) !== -1
  ).pipe(takeUntil(this.destroyed$));
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
  private manageApplicationDrawerRef?: DialogRef;
  private applicationVersionModalRef?: DialogRef;

  private readonly overlay = inject(OverlayService);

  private readonly toast = inject(ToastService);

  ngOnInit() {}

  ngOnDestroy() {
    this.destroyed$.complete();
  }

  openDrawer(templateRef: TemplateRef<unknown>, application?: Application) {
    this.hideDrawer();
    if (application) {
      this.loadApplication(application);
    } else {
      this.reset();
    }
    this.manageApplicationDrawerRef = this.overlay.showDrawer(templateRef);
  }

  hideDrawer() {
    this.manageApplicationDrawerRef?.close();
    this.reset();
  }

  openVersionModal(templateRef: TemplateRef<unknown>, application: Application) {
    this.hideVersionModal();
    this.loadApplication(application);
    this.applicationVersionModalRef = this.overlay.showModal(templateRef);
  }

  hideVersionModal() {
    this.applicationVersionModalRef?.close();
    this.resetVersionForm();
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

  deleteApplication(application: Application) {
    this.overlay
      .confirm(`Really delete ${application.name} and all related deployments?`)
      .pipe(
        filter((result) => result === true),
        switchMap(() => this.applications.delete(application))
      )
      .subscribe();
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
          type: val.type!,
        });
      }
      result.subscribe({
        next: (application) => {
          this.hideDrawer();
          this.toast.success(`${application.name} saved successfully`);
        },
        error: () => this.toast.error(`An error occurred`),
      });
    } else {
      this.editForm.markAllAsTouched();
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
        .subscribe((value) => {
          this.toast.success(`${value.name} created successfully`);
          this.hideVersionModal();
        });
    } else {
      this.newVersionForm.markAllAsTouched();
    }
  }
}
