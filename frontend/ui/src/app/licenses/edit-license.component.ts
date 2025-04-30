import {
  AfterViewInit,
  Component,
  effect,
  ElementRef,
  forwardRef,
  inject,
  Injector,
  OnDestroy,
  OnInit,
  signal,
  ViewChild,
} from '@angular/core';
import {AsyncPipe} from '@angular/common';
import {AutotrimDirective} from '../directives/autotrim.directive';
import {
  ControlValueAccessor,
  FormArray,
  FormBuilder,
  NG_VALUE_ACCESSOR,
  NgControl,
  ReactiveFormsModule,
  TouchedChangeEvent,
  Validators,
} from '@angular/forms';
import {
  faChevronDown,
  faExclamationTriangle,
  faLightbulb,
  faMagnifyingGlass,
  faPen,
  faPlus,
  faXmark,
} from '@fortawesome/free-solid-svg-icons';
import {
  combineLatestWith,
  filter,
  first,
  firstValueFrom,
  map,
  Subject,
  switchMap,
  takeUntil,
  withLatestFrom,
} from 'rxjs';
import {ApplicationLicense} from '../types/application-license';
import {ApplicationsService} from '../services/applications.service';
import {Application, ApplicationVersion} from '@glasskube/distr-sdk';
import {UsersService} from '../services/users.service';
import dayjs from 'dayjs';
import {CdkConnectedOverlay, CdkOverlayOrigin} from '@angular/cdk/overlay';
import {dropdownAnimation} from '../animations/dropdown';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {ArtifactLicense} from '../services/artifact-licenses.service';
import {isArchived} from '../../util/dates';

@Component({
  selector: 'app-edit-license',
  templateUrl: './edit-license.component.html',
  imports: [AsyncPipe, AutotrimDirective, ReactiveFormsModule, CdkOverlayOrigin, CdkConnectedOverlay, FaIconComponent],
  providers: [
    {
      provide: NG_VALUE_ACCESSOR,
      useExisting: forwardRef(() => EditLicenseComponent),
      multi: true,
    },
  ],
  animations: [dropdownAnimation],
})
export class EditLicenseComponent implements OnInit, OnDestroy, AfterViewInit, ControlValueAccessor {
  private injector = inject(Injector);
  private readonly destroyed$ = new Subject<void>();
  private readonly applicationsService = inject(ApplicationsService);
  private readonly usersService = inject(UsersService);
  applications$ = this.applicationsService.list();
  customers$ = this.usersService.getUsers().pipe(
    map((accounts) => accounts.filter((a) => a.userRole === 'customer')),
    first()
  );

  private fb = inject(FormBuilder);
  editForm = this.fb.nonNullable.group({
    id: this.fb.nonNullable.control<string | undefined>(undefined),
    name: this.fb.nonNullable.control<string | undefined>(undefined, Validators.required),
    expiresAt: this.fb.nonNullable.control(''),
    subjectId: this.fb.nonNullable.control<string | undefined>(undefined, Validators.required),
    includeAllItems: this.fb.nonNullable.control<boolean>(true, Validators.required),
    activeVersions: this.fb.array<boolean>([]),
    archivedVersions: this.fb.array<boolean>([]),
    ownerUserAccountId: this.fb.nonNullable.control<string | undefined>(undefined),
    registry: this.fb.nonNullable.group(
      {
        url: this.fb.nonNullable.control(''),
        username: this.fb.nonNullable.control(''),
        password: this.fb.nonNullable.control(''),
      },
      {
        validators: (control) => {
          if (!control.get('url')?.value && !control.get('username')?.value && !control.get('password')?.value) {
            return null;
          }
          if (control.get('url')?.value && control.get('username')?.value && control.get('password')?.value) {
            return null;
          }
          return {
            required: true,
          };
        },
      }
    ),
  });
  editFormLoading = false;
  readonly license = signal<ApplicationLicense | undefined>(undefined);
  readonly selectedSubject = signal<Application | undefined>(undefined);
  readonly includedArchivedVersions = signal<ApplicationVersion[]>([]);

  dropdownOpen = signal(false);
  protected subjectItemsSelected = 0;

  dropdownWidth: number = 0;

  protected readonly faMagnifyingGlass = faMagnifyingGlass;
  protected readonly faChevronDown = faChevronDown;
  protected readonly faPlus = faPlus;
  protected readonly faXmark = faXmark;
  protected readonly faPen = faPen;

  @ViewChild('dropdownTriggerButton') dropdownTriggerButton!: ElementRef<HTMLElement>;

  constructor() {
    effect(() => {
      if (!this.dropdownOpen()) {
        if (
          !this.editForm.controls.includeAllItems.value &&
          !this.editForm.controls.activeVersions.value.some((v) => !!v) &&
          !this.editForm.controls.archivedVersions.getRawValue().some((v) => !!v)
        ) {
          this.editForm.controls.includeAllItems.patchValue(true);
        }
      }
    });
  }

  ngOnInit() {
    this.editForm.controls.includeAllItems.valueChanges.pipe(takeUntil(this.destroyed$)).subscribe((includeAll) => {
      if (includeAll) {
        this.editForm.controls.activeVersions.controls.forEach((c) => c.patchValue(false, {emitEvent: false}));
      }
    });
    this.editForm.controls.activeVersions.valueChanges
      .pipe(takeUntil(this.destroyed$), combineLatestWith(this.editForm.controls.archivedVersions.valueChanges))
      .subscribe(([active, _]) => {
        const archived = this.editForm.controls.archivedVersions.getRawValue();
        if (this.editForm.controls.includeAllItems.value && (active.some((v) => !!v) || archived.some((v) => !!v))) {
          this.editForm.controls.includeAllItems.patchValue(false, {emitEvent: false});
        }
      });
    this.editForm.valueChanges.pipe(takeUntil(this.destroyed$)).subscribe(() => {
      this.onTouched();
      const val = this.editForm.getRawValue();
      if (!val.includeAllItems) {
        this.subjectItemsSelected =
          val.activeVersions.filter((v) => !!v).length + val.archivedVersions.filter((v) => !!v).length;
      }
      if (this.editForm.valid) {
        this.onChange({
          id: val.id,
          name: val.name,
          expiresAt: val.expiresAt ? new Date(val.expiresAt) : undefined,
          applicationId: val.subjectId,
          versions: this.getSelectedVersions(
            val.includeAllItems!,
            val.activeVersions ?? [],
            val.archivedVersions ?? []
          ),
          ownerUserAccountId: val.ownerUserAccountId,
          registryUrl: val.registry.url?.trim() || undefined,
          registryUsername: val.registry.username?.trim() || undefined,
          registryPassword: val.registry.password?.trim() || undefined,
        });
      } else {
        this.onChange(undefined);
      }
    });
    this.editForm.controls.subjectId.valueChanges
      .pipe(
        takeUntil(this.destroyed$),
        switchMap(async (subjectId) => {
          const apps = await firstValueFrom(this.applicationsService.list());
          return apps.find((a) => a.id === subjectId);
        }),
        filter((a) => !!a)
      )
      .subscribe((selectedApp) => {
        const allVersions = selectedApp.versions ?? [];
        const activeVersions = allVersions.filter((v) => !isArchived(v));
        const archivedVersions = allVersions.filter((v) => isArchived(v));

        const appWithResortedVersions = structuredClone(selectedApp);
        appWithResortedVersions.versions = [...activeVersions, ...archivedVersions];

        this.selectedSubject.set(appWithResortedVersions);
        this.activeVersionsArray.clear({emitEvent: activeVersions.length === 0});
        this.archivedVersionsArray.clear({emitEvent: archivedVersions.length === 0});

        const licensedVersions = (this.license() as ApplicationLicense)?.versions;
        let anySelected = false;
        const archivedSelected = [];

        for (let i = 0; i < activeVersions.length; i++) {
          const version = activeVersions[i];
          const selected = !!licensedVersions?.some((v) => v.id === version.id);
          this.activeVersionsArray.push(this.fb.control(selected), {emitEvent: i === activeVersions.length - 1});
          anySelected = anySelected || selected;
        }

        for (let i = 0; i < archivedVersions.length; i++) {
          const version = archivedVersions[i];
          const selected = !!licensedVersions?.some((v) => v.id === version.id);
          if (selected) {
            archivedSelected.push(version);
          }
          const ctrl = this.fb.control(selected);
          ctrl.disable();
          this.archivedVersionsArray.push(ctrl, {emitEvent: i === archivedVersions.length - 1});
          anySelected = anySelected || selected;
        }

        this.includedArchivedVersions.set(archivedSelected);
        if (!anySelected) {
          this.editForm.controls.includeAllItems.patchValue(true);
        }
      });
  }

  selectedApplication(): Application | undefined {
    return this.selectedSubject() as Application;
  }

  private getSelectedVersions(
    includeAllVersions: boolean,
    activeVersionControls: (boolean | null)[],
    archivedVersionControls: (boolean | null)[]
  ): ApplicationVersion[] {
    if (includeAllVersions) {
      return [];
    }
    const app = this.selectedApplication();
    const activeSelected = activeVersionControls
      .map((v, idx) => {
        if (v) {
          return app?.versions?.[idx];
        }
        return undefined;
      })
      .filter((v) => !!v);
    const archivedSelected = archivedVersionControls
      .map((v, idx) => {
        if (v) {
          return app?.versions?.[idx + activeVersionControls.length];
        }
        return undefined;
      })
      .filter((v) => !!v);
    return [...activeSelected, ...archivedSelected];
  }

  ngAfterViewInit() {
    // from https://github.com/angular/angular/issues/45089
    this.injector
      .get(NgControl)
      .control!.events.pipe(takeUntil(this.destroyed$))
      .subscribe((event) => {
        if (event instanceof TouchedChangeEvent) {
          if (event.touched) {
            this.editForm.markAllAsTouched();
          }
        }
      });
  }

  toggleDropdown() {
    this.dropdownOpen.update((v) => !v);
    if (this.dropdownOpen()) {
      this.dropdownWidth = this.dropdownTriggerButton.nativeElement.getBoundingClientRect().width;
    }
  }

  ngOnDestroy() {
    this.destroyed$.next();
    this.destroyed$.complete();
  }

  get activeVersionsArray() {
    return this.editForm.controls.activeVersions as FormArray;
  }

  get archivedVersionsArray() {
    return this.editForm.controls.archivedVersions as FormArray;
  }

  private onChange: (l: ApplicationLicense | ArtifactLicense | undefined) => void = () => {};
  private onTouched: () => void = () => {};

  registerOnChange(fn: (l: ApplicationLicense | ArtifactLicense | undefined) => void): void {
    this.onChange = fn;
  }

  registerOnTouched(fn: any): void {
    this.onTouched = fn;
  }

  writeValue(license: ApplicationLicense | undefined): void {
    this.license.set(license);
    if (license) {
      this.editForm.patchValue({
        id: license.id,
        name: license.name,
        expiresAt: license.expiresAt ? dayjs(license.expiresAt).format('YYYY-MM-DD') : '',
        subjectId: license.applicationId,
        activeVersions: [], // will be set by on-change,
        archivedVersions: [], // will be set by on-change,
        includeAllItems: (license.versions ?? []).length === 0,
        ownerUserAccountId: license.ownerUserAccountId,
        registry: {
          url: license.registryUrl || '',
          username: license.registryUsername || '',
          password: license.registryPassword || '',
        },
      });
      if (license.ownerUserAccountId) {
        this.editForm.controls.subjectId.disable({emitEvent: false});
        this.editForm.controls.ownerUserAccountId.disable({emitEvent: false});
      }
    } else {
      this.editForm.reset();
    }
  }

  protected readonly isArchived = isArchived;
  protected readonly faExclamationTriangle = faExclamationTriangle;
  protected readonly faLightbulb = faLightbulb;
}
