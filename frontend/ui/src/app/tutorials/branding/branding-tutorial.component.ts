import {Component, inject, OnDestroy, OnInit, signal, ViewChild} from '@angular/core';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {
  faArrowRight,
  faB,
  faBox,
  faBoxesStacked,
  faCheck,
  faDownload,
  faLightbulb,
  faPalette,
  faRightToBracket,
} from '@fortawesome/free-solid-svg-icons';
import {CdkStep, CdkStepper, CdkStepperPrevious} from '@angular/cdk/stepper';
import {TutorialStepperComponent} from '../stepper/tutorial-stepper.component';
import {OrganizationBrandingService} from '../../services/organization-branding.service';
import {RouterLink} from '@angular/router';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {AutotrimDirective} from "../../directives/autotrim.directive";
import {faCircleCheck} from "@fortawesome/free-regular-svg-icons";
import {firstValueFrom, lastValueFrom, map, Observable, Subject, takeUntil, tap} from "rxjs";
import {OrganizationBranding} from "../../../../../../sdk/js/src";
import {base64ToBlob} from "../../../util/blob";
import {getFormDisplayedError} from "../../../util/errors";
import {HttpErrorResponse} from "@angular/common/http";
import {ToastService} from "../../services/toast.service";
import {UsersService} from '../../services/users.service';

const defaultBrandingDescription = `# Welcome

In this Customer Portal you can manage your deployments.
`

@Component({
  selector: 'app-branding-tutorial',
  imports: [
    ReactiveFormsModule,
    CdkStep,
    TutorialStepperComponent,
    RouterLink,
    FaIconComponent,
    CdkStepperPrevious,
    AutotrimDirective,
  ],
  templateUrl: './branding-tutorial.component.html',
})
export class BrandingTutorialComponent implements OnInit, OnDestroy {
  loading = signal(true);
  private readonly destroyed$ = new Subject<void>();
  protected readonly faBox = faBox;
  protected readonly faDownload = faDownload;
  protected readonly faPalette = faPalette;
  protected readonly faBoxesStacked = faBoxesStacked;
  protected readonly faB = faB;
  protected readonly faLightbulb = faLightbulb;
  @ViewChild('stepper') private stepper!: CdkStepper;
  private toast = inject(ToastService);
  protected readonly brandingService = inject(OrganizationBrandingService);
  protected readonly usersService = inject(UsersService);
  private organizationBranding?: OrganizationBranding;
  protected readonly welcomeFormGroup = new FormGroup({});
  protected readonly brandingFormGroup = new FormGroup({
    titleDone: new FormControl<boolean>(false),
    title: new FormControl<string>('', {nonNullable: true, validators: Validators.required}),
    descriptionDone: new FormControl<boolean>(false),
    description: new FormControl<string>('', {nonNullable: true, validators: Validators.required}),
  });
  protected readonly inviteFormGroup = new FormGroup({
    customerEmail: new FormControl<string>('', {nonNullable: true, validators: Validators.required}),
    inviteDone: new FormControl<boolean>(false),
    customerConfirmed: new FormControl<boolean>(false, {nonNullable: true}),
  });

  // TODO on load, check existing tutorial state and also check if branding already exists and fill form accordingly
  // (customer invite probably can't be checked, because even if one exists, they could have been invited by somebody else)

  async ngOnInit() {
    this.brandingFormGroup.controls.title.statusChanges.pipe(takeUntil(this.destroyed$)).subscribe(status => {
      this.brandingFormGroup.controls.titleDone.patchValue(status !== 'INVALID');
    })
    this.brandingFormGroup.controls.description.statusChanges.pipe(takeUntil(this.destroyed$)).subscribe(status => {
      this.brandingFormGroup.controls.descriptionDone.patchValue(status !== 'INVALID');
    })

    try {
      this.organizationBranding = await lastValueFrom(this.brandingService.get());
      this.brandingFormGroup.patchValue({
        title: this.organizationBranding.title,
        description: this.organizationBranding.description || defaultBrandingDescription,
      });
    } catch (e) {
      const msg = getFormDisplayedError(e);
      if (msg && e instanceof HttpErrorResponse && e.status !== 404) {
        // it's a valid use case for an organization to have no branding (hence 404 is not shown in toast)
        this.toast.error(msg);
      }
    } finally {
      this.loading.set(false);
    }
  }

  protected continueFromWelcome() {
    // TODO put tutorial state
    this.brandingService.get();
    this.stepper.next();
  }

  protected back() {
    // const oldStep = this.stepper.selected!;
    // const wasCompleted = oldStep.completed;
    this.stepper.previous(); // why does this set completed to true if its not submitting ????
    /*if (!wasCompleted) {
      oldStep.completed = false;
    }*/
  }

  protected async continueFromBranding() {
    this.brandingFormGroup.markAllAsTouched();
    if (this.brandingFormGroup.valid) {
      // TODO put tutorial state
      // this.stepper.selected!.completed = true;

      this.loading.set(true);
      const formData = new FormData();
      formData.set('title', this.brandingFormGroup.value.title ?? '');
      // formData.set('description', this.form.value.description ?? '');

      const id = this.organizationBranding?.id;
      let req: Observable<OrganizationBranding>;
      if (id) {
        req = this.brandingService.update(formData);
      } else {
        req = this.brandingService.create(formData);
      }

      try {
        this.organizationBranding = await lastValueFrom(req);
        this.stepper.next();
      } catch (e) {
        const msg = getFormDisplayedError(e);
        if (msg) {
          this.toast.error(msg);
        }
      } finally {
        this.loading.set(false);
      }
    }
  }

  protected async sendInviteMail() {
    this.inviteFormGroup.markAllAsTouched();
    if(this.inviteFormGroup.valid) {
      this.loading.set(true);
      try {
        const result = await firstValueFrom(
          this.usersService.addUser({
            email: this.inviteFormGroup.value.customerEmail!,
            userRole: 'customer',
          })
        );
        this.inviteFormGroup.controls.inviteDone.patchValue(true);
        // TODO put tutorial state
        // this.inviteUrl = result.inviteUrl;
      } catch (e) {
        const msg = getFormDisplayedError(e);
        if (msg) {
          this.toast.error(msg);
        }
      } finally {
        this.loading.set(false);
      }
    }
  }

  protected completeAndExit() {
    this.brandingFormGroup.markAllAsTouched();
    if (this.brandingFormGroup.valid) {
      // TODO put tutorial state
      this.stepper.selected!.completed = true;
    }
  }

  protected readonly faArrowRight = faArrowRight;
  protected readonly faRightToBracket = faRightToBracket;
  protected readonly faCheck = faCheck;
  protected readonly faCircleCheck = faCircleCheck;

  ngOnDestroy() {
    this.destroyed$.next();
    this.destroyed$.complete();
  }
}
