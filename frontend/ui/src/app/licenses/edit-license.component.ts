import {
  AfterViewInit,
  Component,
  effect,
  ElementRef,
  forwardRef,
  inject,
  Injector, Input,
  OnDestroy,
  OnInit,
  signal,
  ViewChild,
  WritableSignal,
} from '@angular/core';
import {AsyncPipe} from '@angular/common';
import {AutotrimDirective} from '../directives/autotrim.directive';
import {
  ControlValueAccessor,
  FormArray,
  FormBuilder,
  FormControl,
  FormGroup,
  NG_VALUE_ACCESSOR,
  NgControl,
  ReactiveFormsModule,
  TouchedChangeEvent,
  Validators,
} from '@angular/forms';
import {faChevronDown, faMagnifyingGlass, faPen, faPlus, faXmark} from '@fortawesome/free-solid-svg-icons';
import {first, firstValueFrom, map, Subject, switchMap, takeUntil} from 'rxjs';
import {ApplicationLicense} from '../types/application-license';
import {ApplicationsService} from '../services/applications.service';
import {Application, ApplicationVersion, BaseModel, Named, UserAccount} from '@glasskube/distr-sdk';
import {UsersService} from '../services/users.service';
import dayjs from 'dayjs';
import {CdkConnectedOverlay, CdkOverlayOrigin} from '@angular/cdk/overlay';
import {dropdownAnimation} from '../animations/dropdown';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {ArtifactLicense} from '../services/artifact-licenses.service';
import {Artifact, ArtifactsService, ArtifactTag, ArtifactWithTags} from '../services/artifacts.service';

export type LicenseType = 'application' | 'artifact';

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
  @Input({required: true}) licenseType!: LicenseType;
  private injector = inject(Injector);
  private readonly destroyed$ = new Subject<void>();
  private readonly applicationsService = inject(ApplicationsService);
  private readonly artifactsService = inject(ArtifactsService);
  private readonly usersService = inject(UsersService);
  applications$ = this.applicationsService.list();
  artifacts$ = this.artifactsService.list();
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
    subjectItems: this.fb.array<boolean>([]),
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
  readonly license = signal<ApplicationLicense | ArtifactLicense | undefined>(undefined);
  readonly selectedSubject = signal<Application | ArtifactWithTags | undefined>(undefined);

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
          !this.editForm.controls.subjectItems.value.some((v) => !!v)
        ) {
          this.editForm.controls.includeAllItems.patchValue(true);
        }
      }
    });
  }

  ngOnInit() {
    this.editForm.controls.includeAllItems.valueChanges.pipe(takeUntil(this.destroyed$)).subscribe((includeAll) => {
      if (includeAll) {
        this.editForm.controls.subjectItems.controls.forEach((c) => c.patchValue(false, {emitEvent: false}));
      }
    });
    this.editForm.controls.subjectItems.valueChanges.pipe(takeUntil(this.destroyed$)).subscribe((val) => {
      if (this.editForm.controls.includeAllItems.value && val.some((v) => !!v)) {
        this.editForm.controls.includeAllItems.patchValue(false, {emitEvent: false});
      }
    });
    this.editForm.valueChanges.pipe(takeUntil(this.destroyed$)).subscribe(() => {
      this.onTouched();
      const val = this.editForm.getRawValue();
      if (!val.includeAllItems) {
        this.subjectItemsSelected = val.subjectItems.filter((v) => !!v).length;
      }
      if (this.editForm.valid) {
        this.onChange({
          id: val.id,
          name: val.name,
          expiresAt: val.expiresAt ? new Date(val.expiresAt) : undefined,
          applicationId: this.licenseType === 'application' ? val.subjectId : undefined,
          artifactId: this.licenseType === 'artifact' ? val.subjectId : undefined,
          versions: this.licenseType === 'application' ? this.getSelectedVersions(val.includeAllItems!, val.subjectItems ?? []) : undefined,
          artifactTags: this.licenseType === 'artifact' ? this.getSelectedTags(val.includeAllItems!, val.subjectItems ?? []) : undefined,
          ownerUserAccountId: val.ownerUserAccountId,
          registryUrl: this.licenseType === 'application' ? (val.registry.url?.trim() || undefined) : undefined,
          registryUsername: this.licenseType === 'application' ? (val.registry.username?.trim() || undefined) : undefined,
          registryPassword: this.licenseType === 'application' ? (val.registry.password?.trim() || undefined) : undefined,
        });
      } else {
        this.onChange(undefined);
      }
    });
    this.editForm.controls.subjectId.valueChanges
      .pipe(
        takeUntil(this.destroyed$),
        switchMap(async (subjectId) => {
          if(this.licenseType === 'application') {
            const apps = await firstValueFrom(this.applicationsService.list());
            return apps.find((a) => a.id === subjectId);
          } else {
            const artifacts = await firstValueFrom(this.artifactsService.list());
            return artifacts.find((a) => a.id === subjectId);
          }
        })
      )
      .subscribe((selectedSubject) => {
        this.subjectItemsArray.clear({emitEvent: false});
        const allItems = this.licenseType === 'application' ?
          (selectedSubject as Application)?.versions ?? [] :
          (selectedSubject as ArtifactWithTags)?.tags ?? [];
        const licenseItems = this.licenseType === 'application' ?
          (this.license() as ApplicationLicense)?.versions :
          (this.license() as ArtifactLicense)?.artifactTags;
        let anySelected = false;
        for (let i = 0; i < allItems.length; i++) {
          const item = allItems[i];
          const selected = !!licenseItems?.some((v) => v.id === item.id);
          this.subjectItemsArray.push(this.fb.control(selected), {emitEvent: i === allItems.length - 1});
          anySelected = anySelected || selected;
        }
        if (!anySelected) {
          this.editForm.controls.includeAllItems.patchValue(true);
        }
        this.selectedSubject.set(selectedSubject);
      });
  }

  selectedApplication(): Application | undefined {
    return this.selectedSubject() as Application;
  }

  private getSelectedVersions(includeAllVersions: boolean, itemControls: (boolean | null)[]): ApplicationVersion[] {
    if (includeAllVersions) {
      return [];
    }
    const app = this.selectedApplication();
    return itemControls
      .map((v, idx) => {
        if (v) {
          return app?.versions?.[idx];
        }
        return undefined;
      })
      .filter((v) => !!v);
  }

  selectedArtifact(): ArtifactWithTags | undefined {
    return this.selectedSubject() as ArtifactWithTags;
  }

  private getSelectedTags(includeAllVersions: boolean, itemControls: (boolean | null)[]): ArtifactTag[] {
    if (includeAllVersions) {
      return [];
    }
    const artifact = this.selectedArtifact();
    return itemControls
      .map((v, idx) => {
        if (v) {
          return artifact?.tags?.[idx];
        }
        return undefined;
      })
      .filter((v) => !!v);
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

  get subjectItemsArray() {
    return this.editForm.controls.subjectItems as FormArray;
  }

  private onChange: (l: ApplicationLicense | ArtifactLicense | undefined) => void = () => {};
  private onTouched: () => void = () => {};

  registerOnChange(fn: (l: ApplicationLicense | ArtifactLicense | undefined) => void): void {
    this.onChange = fn;
  }

  registerOnTouched(fn: any): void {
    this.onTouched = fn;
  }

  writeValue(license: ApplicationLicense | ArtifactLicense | undefined): void {
    this.license.set(license);
    if (license) {
      const {subjectId, includeAllItems, registry} = this.getLicenseTypeSpecificValuesForForm(license);
      this.editForm.patchValue({
        id: license.id,
        name: license.name,
        expiresAt: license.expiresAt ? dayjs(license.expiresAt).format('YYYY-MM-DD') : '',
        subjectId,
        subjectItems: [], // will be set by on-change,
        includeAllItems,
        ownerUserAccountId: license.ownerUserAccountId,
        registry
      });
      if (license.ownerUserAccountId) {
        this.editForm.controls.subjectId.disable({emitEvent: false});
        this.editForm.controls.ownerUserAccountId.disable({emitEvent: false});
      }
    } else {
      this.editForm.reset();
    }
  }

  private getLicenseTypeSpecificValuesForForm(license: ApplicationLicense | ArtifactLicense): {
    subjectId?: string;
    includeAllItems: boolean;
    registry?: {
      url?: string;
      username?: string;
      password?: string;
    }
  } {
    if(this.licenseType === 'application') {
      const appLicense = license as ApplicationLicense;
      return {
        subjectId: appLicense.applicationId,
        includeAllItems: (appLicense.versions ?? []).length === 0,
        registry: {
          url: appLicense.registryUrl || '',
          username: appLicense.registryUsername || '',
          password: appLicense.registryPassword || '',
        }
      }
    } else {
      const artifactLicense = license as ArtifactLicense;
      return {
        subjectId: artifactLicense.artifactId,
        includeAllItems: (artifactLicense.artifactTags ?? []).length === 0
      }
    }
  }

}
