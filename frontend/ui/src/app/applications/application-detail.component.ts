import {GlobalPositionStrategy, OverlayModule} from '@angular/cdk/overlay';
import {AsyncPipe, DatePipe, NgOptimizedImage} from '@angular/common';
import {Component, ElementRef, inject, OnDestroy, OnInit, signal, TemplateRef, ViewChild} from '@angular/core';
import {FormControl, FormGroup, FormsModule, ReactiveFormsModule, Validators} from '@angular/forms';
import {ActivatedRoute, Router, RouterLink} from '@angular/router';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {
  faArchive,
  faBox,
  faBoxesStacked,
  faCheck,
  faChevronDown,
  faEdit, faEye,
  faMagnifyingGlass,
  faTrash,
  faXmark,
} from '@fortawesome/free-solid-svg-icons';
import {Application, ApplicationVersion, HelmChartType} from '@glasskube/distr-sdk';
import {
  catchError,
  combineLatest,
  combineLatestAll,
  combineLatestWith,
  distinctUntilChanged,
  EMPTY,
  filter,
  firstValueFrom,
  lastValueFrom,
  map,
  Observable,
  startWith,
  Subject,
  switchMap,
  take,
  takeUntil,
  tap,
} from 'rxjs';
import {isArchived} from '../../util/dates';
import {getFormDisplayedError} from '../../util/errors';
import {disableControlsWithoutEvent, enableControlsWithoutEvent} from '../../util/forms';
import {SecureImagePipe} from '../../util/secureImage';
import {dropdownAnimation} from '../animations/dropdown';
import {EditorComponent} from '../components/editor.component';
import {UuidComponent} from '../components/uuid';
import {AutotrimDirective} from '../directives/autotrim.directive';
import {ApplicationsService} from '../services/applications.service';
import {DialogRef, OverlayService} from '../services/overlay.service';
import {ToastService} from '../services/toast.service';
import {DeploymentModalComponent} from '../deployments/deployment-modal.component';
import {DeploymentFormComponent} from '../deployment-form/deployment-form.component';

@Component({
  selector: 'app-application-detail',
  imports: [
    ReactiveFormsModule,
    OverlayModule,
    AsyncPipe,
    RouterLink,
    FaIconComponent,
    NgOptimizedImage,
    UuidComponent,
    AutotrimDirective,
    DatePipe,
    EditorComponent,
    SecureImagePipe,
    FormsModule,
  ],
  templateUrl: './application-detail.component.html',
  animations: [dropdownAnimation],
})
export class ApplicationDetailComponent implements OnInit, OnDestroy {
  private readonly destroyed$ = new Subject<void>();
  private readonly toast = inject(ToastService);
  private readonly overlay = inject(OverlayService);
  private readonly applicationService = inject(ApplicationsService);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  readonly applications$: Observable<Application[]> = this.applicationService.list();
  filterForm = new FormGroup({
    showArchived: new FormControl<boolean>(false),
  });
  readonly application$: Observable<Application | undefined> = this.route.paramMap.pipe(
    map((params) => params.get('applicationId')?.trim()),
    distinctUntilChanged(),
    tap(() => {
      this.newVersionForm.reset();
      this.newVersionFormLoading.set(false);
    }),
    combineLatestWith(this.applications$),
    map(([id, applications]) => applications.find((a) => a.id === id)),
    tap((app) => {
      this.editForm.disable();
      if (app) {
        this.editForm.patchValue({name: app.name});
        this.enableTypeSpecificGroups(app);
      }
    })
  );
  readonly visibleVersions$ = this.application$.pipe(
    combineLatestWith(this.filterForm.valueChanges.pipe(startWith({showArchived: false}))),
    map(([app, filter]) => {
      if (app && !filter.showArchived) {
        return (app.versions ?? []).filter((av) => !isArchived(av));
      }
      return app?.versions;
    })
  );

  protected selectedVersionIds = signal(new Set<string>());

  newVersionForm = new FormGroup({
    versionName: new FormControl('', Validators.required),
    kubernetes: new FormGroup(
      {
        chartType: new FormControl<HelmChartType>('repository', {
          nonNullable: true,
          validators: Validators.required,
        }),
        chartName: new FormControl('', Validators.required),
        chartUrl: new FormControl('', Validators.required),
        chartVersion: new FormControl('', Validators.required),
        baseValues: new FormControl(''),
        template: new FormControl(''),
      },
      (v) => {
        if (v.value.chartType === 'oci' && v.value.chartUrl && !/^oci:\/\/.+/.test(v.value.chartUrl)) {
          return {chartUrlOci: true};
        }
        if (v.value.chartType === 'repository' && v.value.chartUrl && !/^https:\/\/.+/.test(v.value.chartUrl)) {
          return {chartUrlHttps: true};
        }
        return null;
      }
    ),
    docker: new FormGroup({
      compose: new FormControl(''),
      template: new FormControl(''),
    }),
  });
  newVersionFormLoading = signal(false);
  editForm = new FormGroup({
    name: new FormControl('', Validators.required),
  });
  editFormLoading = signal(false);

  @ViewChild('versionDetailsModal') protected readonly versionDetailsModal!: TemplateRef<unknown>;
  protected readonly selectedVersion = signal<ApplicationVersion | undefined>(undefined);
  versionDetailsForm = new FormGroup({
    kubernetes: new FormGroup(
      {
        baseValues: new FormControl(''),
        template: new FormControl(''),
      }
    ),
    docker: new FormGroup({
      compose: new FormControl(''),
      template: new FormControl(''),
    }),
  })

  protected readonly faBoxesStacked = faBoxesStacked;
  protected readonly faChevronDown = faChevronDown;
  protected readonly faEdit = faEdit;
  protected readonly faCheck = faCheck;
  protected readonly faXmark = faXmark;
  protected readonly faTrash = faTrash;
  protected readonly faArchive = faArchive;
  protected readonly faMagnifyingGlass = faMagnifyingGlass;
  protected readonly faBox = faBox;
  protected readonly isArchived = isArchived;
  readonly breadcrumbDropdown = signal(false);
  readonly isVersionFormExpanded = signal(false);
  breadcrumbDropdownWidth: number = 0;
  @ViewChild('dropdownTriggerButton') dropdownTriggerButton!: ElementRef<HTMLElement>;
  @ViewChild('nameInput') nameInputElem?: ElementRef<HTMLInputElement>;
  private modal?: DialogRef;

  ngOnInit() {
    this.route.url.subscribe(() => this.breadcrumbDropdown.set(false));

    this.newVersionForm.controls.kubernetes.controls.chartType.valueChanges
      .pipe(takeUntil(this.destroyed$))
      .subscribe((type) => {
        if (type === 'repository') {
          this.newVersionForm.controls.kubernetes.controls.chartName.enable();
        } else {
          this.newVersionForm.controls.kubernetes.controls.chartName.disable();
        }
      });
    this.versionDetailsForm.disable();
  }

  ngOnDestroy() {
    this.destroyed$.next();
    this.destroyed$.complete();
  }

  toggleBreadcrumbDropdown() {
    this.breadcrumbDropdown.update((v) => !v);
    if (this.breadcrumbDropdown()) {
      this.breadcrumbDropdownWidth = this.dropdownTriggerButton.nativeElement.getBoundingClientRect().width;
    }
  }

  enableApplicationEdit(application: Application) {
    this.editForm.enable();
    this.editForm.patchValue({name: application.name});
    setTimeout(() => this.nameInputElem?.nativeElement.focus(), 10);
  }

  cancelApplicationEdit() {
    this.editForm.disable();
  }

  async saveApplication(application: Application) {
    if (this.editForm.valid) {
      this.editFormLoading.set(true);
      try {
        await lastValueFrom(
          this.applicationService.update({
            ...application,
            name: this.editForm.value.name!.trim(),
          })
        );
      } catch (e) {
        const msg = getFormDisplayedError(e);
        if (msg) {
          this.toast.error(msg);
        }
      } finally {
        this.editFormLoading.set(false);
      }
    }
  }

  async createVersion(application: Application) {
    this.newVersionForm.markAllAsTouched();
    if (this.newVersionForm.valid && application) {
      this.newVersionFormLoading.set(true);
      let res;
      if (application.type === 'docker') {
        res = this.applicationService.createApplicationVersionForDocker(
          application,
          {name: this.newVersionForm.controls.versionName.value!},
          this.newVersionForm.controls.docker.controls.compose.value!,
          this.newVersionForm.controls.docker.controls.template.value
        );
      } else {
        const versionFormVal = this.newVersionForm.controls.kubernetes.value;
        const version = {
          name: this.newVersionForm.controls.versionName.value!,
          chartType: versionFormVal.chartType!,
          chartName: versionFormVal.chartName ?? undefined,
          chartUrl: versionFormVal.chartUrl!,
          chartVersion: versionFormVal.chartVersion!,
        };
        res = this.applicationService.createApplicationVersionForKubernetes(
          application,
          version,
          versionFormVal.baseValues,
          versionFormVal.template
        );
      }

      try {
        const av = await firstValueFrom(res);
        this.toast.success(`${av.name} created successfully`);
        this.newVersionForm.reset();
        this.enableTypeSpecificGroups(application);
      } catch (e) {
        const msg = getFormDisplayedError(e);
        if (msg) {
          this.toast.error(msg);
        }
      } finally {
        this.newVersionFormLoading.set(false);
      }
    }
  }

  async fillVersionFormWith(application: Application, version: ApplicationVersion) {
    this.isVersionFormExpanded.set(true);
    const val = await this.loadVersionDetails(application, version);
    if(val) {
      this.newVersionForm.patchValue(val);
    }
  }

  async loadVersionDetails(application: Application, version: ApplicationVersion) {
    if (application.type === 'kubernetes') {
      try {
        const template = await firstValueFrom(this.applicationService.getTemplateFile(application.id!, version.id!));
        const values = await firstValueFrom(this.applicationService.getValuesFile(application.id!, version.id!));
        return {
          kubernetes: {
            chartType: version.chartType,
            chartName: version.chartName,
            chartUrl: version.chartUrl,
            chartVersion: version.chartVersion,
            baseValues: values,
            template: template,
          },
        };
      } catch (e) {
        const msg = getFormDisplayedError(e);
        if (msg) {
          this.toast.error(msg);
        }
        return undefined;
      }
    } else if (application.type === 'docker') {
      try {
        const template = await firstValueFrom(this.applicationService.getTemplateFile(application.id!, version.id!));
        const compose = await firstValueFrom(this.applicationService.getComposeFile(application.id!, version.id!));
        return {
          docker: {
            compose,
            template: template,
          },
        };
      } catch (e) {
        const msg = getFormDisplayedError(e);
        if (msg) {
          this.toast.error(msg);
        }
        return undefined;
      }
    }
    return undefined;
  }

  deleteApplication(application: Application) {
    this.overlay
      .confirm(`Really delete ${application.name}?`)
      .pipe(
        filter((result) => result === true),
        switchMap(async () => {
          await lastValueFrom(this.applicationService.delete(application));
          return this.router.navigate(['/applications']);
        }),
        catchError((e) => {
          const msg = getFormDisplayedError(e);
          if (msg) {
            this.toast.error(msg);
          }
          return EMPTY;
        })
      )
      .subscribe();
  }

  archiveVersion(application: Application, version: ApplicationVersion) {
    this.overlay
      .confirm(`Really archive ${version.name}? Existing deployments will continue to work.`)
      .pipe(
        filter((result) => result === true),
        switchMap(() =>
          this.applicationService.updateApplicationVersion(application, {
            ...version,
            archivedAt: new Date().toISOString(),
          })
        ),
        catchError((e) => {
          const msg = getFormDisplayedError(e);
          if (msg) {
            this.toast.error(msg);
          }
          return EMPTY;
        })
      )
      .subscribe();
  }

  async bulkArchiveVersions(app: Application) {
    this.overlay
      .confirm(`Really archive ${this.selectedVersionIds().size} versions? Existing deployments will continue to work.`)
      .pipe(
        filter((it) => it === true),
        switchMap(() =>
          this.applicationService.patch(app.id!, {
            versions: app?.versions
              ?.filter((version) => this.isVersionSelected(version))
              .map((version) => ({id: version.id!, archivedAt: new Date().toISOString()})),
          })
        )
      )
      .subscribe({
        next: () => {
          this.toast.success(`${this.selectedVersionIds().size} archived`);
          this.selectedVersionIds.set(new Set());
        },
        error: (e) => {
          const msg = getFormDisplayedError(e);
          if (msg) {
            this.toast.error(msg);
          }
        },
      });
  }

  async unArchiveVersion(application: Application, version: ApplicationVersion) {
    try {
      await lastValueFrom(
        this.applicationService.updateApplicationVersion(application, {
          ...version,
          archivedAt: undefined,
        })
      );
    } catch (e) {
      const msg = getFormDisplayedError(e);
      if (msg) {
        this.toast.error(msg);
      }
    }
  }

  protected hideModal(): void {
    this.versionDetailsForm.reset();
    this.modal?.close();
  }

  async openVersionDetails(application: Application, version: ApplicationVersion) {
    this.hideModal();
    this.selectedVersion.set(version);
    const val = await this.loadVersionDetails(application, version);
    if(val) {
      this.versionDetailsForm.patchValue(val);
      this.modal = this.overlay.showModal(this.versionDetailsModal, {
        positionStrategy: new GlobalPositionStrategy().centerHorizontally().centerVertically(),
      });
    }
  }

  private enableTypeSpecificGroups(app: Application) {
    if (app.type === 'kubernetes') {
      enableControlsWithoutEvent(this.newVersionForm.controls.kubernetes);
      disableControlsWithoutEvent(this.newVersionForm.controls.docker);
    } else {
      enableControlsWithoutEvent(this.newVersionForm.controls.docker);
      disableControlsWithoutEvent(this.newVersionForm.controls.kubernetes);
    }
  }

  public async uploadImage(data: Application) {
    const fileId = await firstValueFrom(this.overlay.uploadImage({imageUrl: data.imageUrl}));
    if (!fileId || data.imageUrl?.includes(fileId)) {
      return;
    }
    await firstValueFrom(this.applicationService.patchImage(data.id!, fileId));
  }

  protected toggleVersionSelected(version: ApplicationVersion): void {
    const id = version.id;
    if (id === undefined) {
      return;
    }

    this.selectedVersionIds.update((ids) => {
      if (this.isVersionSelected(version)) {
        ids.delete(id);
        return ids;
      } else {
        return ids.add(id);
      }
    });
  }

  protected isVersionSelected(version: ApplicationVersion): boolean {
    return version.id !== undefined && this.selectedVersionIds().has(version.id);
  }

  protected toggleAllVersionsSelected(versions: ApplicationVersion[]) {
    versions = versions.filter((version) => !isArchived(version));
    if (this.isAllVersionsSelected(versions)) {
      this.selectedVersionIds.set(new Set());
    } else {
      this.selectedVersionIds.set(new Set(versions.map((v) => v.id).filter((id) => id !== undefined)));
    }
  }

  protected isAllVersionsSelected(versions: ApplicationVersion[]): boolean {
    versions = versions.filter((version) => !isArchived(version));
    return (
      this.selectedVersionIds().size > 0 &&
      versions.length === this.selectedVersionIds().size &&
      versions.every((version) => this.isVersionSelected(version))
    );
  }

  protected readonly faEye = faEye;
}
