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
import {Router, RouterLink} from '@angular/router';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {AutotrimDirective} from '../../directives/autotrim.directive';
import {faCircleCheck} from '@fortawesome/free-regular-svg-icons';
import {firstValueFrom, lastValueFrom, map, Observable, Subject, takeUntil, tap} from 'rxjs';
import {OrganizationBranding} from '../../../../../../sdk/js/src';
import {base64ToBlob} from '../../../util/blob';
import {getFormDisplayedError} from '../../../util/errors';
import {HttpErrorResponse} from '@angular/common/http';
import {ToastService} from '../../services/toast.service';
import {UsersService} from '../../services/users.service';
import {TutorialsService} from '../../services/tutorials.service';
import {TutorialProgress} from '../../types/tutorials';

const defaultBrandingDescription = `# Welcome

In this Customer Portal you can manage your deployments.
`;

const tutorialId = 'branding';
const welcomeStep = 'welcome';
const welcomeTaskStart = 'start';
const brandingStep = 'branding';
const brandingTaskSet = 'set';
const customerStep = 'customer';
const customerTaskInvite = 'invite';
const customerTaskLogin = 'login';

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
  private readonly router = inject(Router);
  protected readonly toast = inject(ToastService);
  protected readonly brandingService = inject(OrganizationBrandingService);
  protected readonly usersService = inject(UsersService);
  protected readonly tutorialsService = inject(TutorialsService);
  protected progress?: TutorialProgress;
  private organizationBranding?: OrganizationBranding;
  protected readonly welcomeFormGroup = new FormGroup({});
  protected readonly brandingFormGroup = new FormGroup({
    titleDone: new FormControl<boolean>(false, Validators.requiredTrue),
    title: new FormControl<string>('', {nonNullable: true, validators: Validators.required}),
    descriptionDone: new FormControl<boolean>(false, Validators.requiredTrue),
    description: new FormControl<string>('', {nonNullable: true, validators: Validators.required}),
  });
  protected readonly inviteFormGroup = new FormGroup({
    customerEmail: new FormControl<string>('', {nonNullable: true, validators: Validators.required}),
    inviteDone: new FormControl<boolean>(false, Validators.requiredTrue),
    customerConfirmed: new FormControl<boolean>(false, Validators.requiredTrue),
  });

  async ngOnInit() {
    try {
      this.progress = await lastValueFrom(this.tutorialsService.get(tutorialId));
    } catch (e) {
      const msg = getFormDisplayedError(e);
      if (msg && e instanceof HttpErrorResponse && e.status !== 404) {
        // it's a valid use case for a tutorial progress not to exist yet
        this.toast.error(msg);
      }
    }
  }

  protected async continueFromWelcome() {
    // prepare branding step
    try {
      this.organizationBranding = await lastValueFrom(this.brandingService.get());
      this.brandingFormGroup.patchValue({
        title: this.organizationBranding.title,
        titleDone: !!this.organizationBranding.title,
        description: this.organizationBranding.description || defaultBrandingDescription,
        descriptionDone: !!this.organizationBranding.description,
      });
      if (this.organizationBranding.title) {
        this.brandingFormGroup.controls.title.disable();
        this.brandingFormGroup.controls.titleDone.disable();
      }
      if (this.organizationBranding.description) {
        this.brandingFormGroup.controls.description.disable();
        this.brandingFormGroup.controls.descriptionDone.disable();
      }
    } catch (e) {
      const msg = getFormDisplayedError(e);
      if (msg && e instanceof HttpErrorResponse && e.status !== 404) {
        // it's a valid use case for an organization to have no branding (hence 404 is not shown in toast)
        this.toast.error(msg);
      }
    } finally {
      this.loading.set(false);
    }

    if (!this.progress) {
      this.loading.set(true);
      try {
        this.progress = await lastValueFrom(
          this.tutorialsService.save(tutorialId, {
            stepId: welcomeStep,
            taskId: welcomeTaskStart,
          })
        );
        this.stepper.next();
      } catch (e) {
        const msg = getFormDisplayedError(e);
        if (msg) {
          this.toast.error(msg);
        }
      } finally {
        this.loading.set(false);
      }
    } else {
      this.stepper.next();
    }
  }

  protected async continueFromBranding() {
    this.brandingFormGroup.markAllAsTouched();
    if (this.brandingFormGroup.valid) {
      this.loading.set(true);
      const formVal = this.brandingFormGroup.getRawValue();
      const formData = new FormData();
      formData.set('title', formVal.title);
      formData.set('description', formVal.description);

      const id = this.organizationBranding?.id;
      let req: Observable<OrganizationBranding>;
      if (id) {
        req = this.brandingService.update(formData);
      } else {
        req = this.brandingService.create(formData);
      }

      try {
        this.organizationBranding = await lastValueFrom(req);
        this.progress = await lastValueFrom(
          this.tutorialsService.save(tutorialId, {
            stepId: brandingStep,
            taskId: brandingTaskSet,
          })
        );
        this.brandingFormGroup.controls.title.disable();
        this.brandingFormGroup.controls.titleDone.patchValue(true);
        this.brandingFormGroup.controls.titleDone.disable();
        this.brandingFormGroup.controls.description.disable();
        this.brandingFormGroup.controls.descriptionDone.patchValue(true);
        this.brandingFormGroup.controls.descriptionDone.disable();

        this.prepareCustomerStep();
        this.stepper.next();
      } catch (e) {
        const msg = getFormDisplayedError(e);
        if (msg) {
          this.toast.error(msg);
        }
      } finally {
        this.loading.set(false);
      }
    } else if (this.brandingFormGroup.disabled) {
      // when all inputs are disabled, the form group is disabled too, i.e. branding has already existed before
      this.prepareCustomerStep();
      this.stepper.selected!.completed = true; // because completed is only true automatically, if form group status is valid
      this.stepper.next();
    }
  }

  private prepareCustomerStep() {
    // prepare the email form
    const email = (this.progress?.events ?? []).find(
      (e) => e.stepId === customerStep && e.taskId === customerTaskInvite
    )?.value;
    if (email && typeof email === 'string') {
      this.inviteFormGroup.controls.customerEmail.patchValue(email);
      this.inviteFormGroup.controls.inviteDone.patchValue(true);
      this.inviteFormGroup.controls.customerEmail.disable();
      this.inviteFormGroup.controls.inviteDone.disable();
    }

    const login = (this.progress?.events ?? []).find(
      (e) => e.stepId === customerStep && e.taskId === customerTaskLogin
    );
    if (login) {
      this.inviteFormGroup.controls.customerConfirmed.patchValue(true);
      this.inviteFormGroup.controls.customerConfirmed.disable();
    }
  }

  protected async sendInviteMail() {
    this.inviteFormGroup.markAllAsTouched();
    if (this.inviteFormGroup.controls.customerEmail.valid) {
      this.loading.set(true);
      try {
        const email = this.inviteFormGroup.value.customerEmail!;
        await lastValueFrom(
          this.usersService.addUser({
            email,
            userRole: 'customer',
          })
        );
        this.inviteFormGroup.controls.inviteDone.patchValue(true);
        this.progress = await lastValueFrom(
          this.tutorialsService.save(tutorialId, {
            stepId: customerStep,
            taskId: customerTaskInvite,
            value: email,
          })
        );
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

  protected async completeAndExit() {
    this.inviteFormGroup.markAllAsTouched();
    if (this.inviteFormGroup.valid) {
      this.loading.set(true);
      this.progress = await lastValueFrom(
        this.tutorialsService.save(tutorialId, {
          stepId: customerStep,
          taskId: customerTaskLogin,
          markCompleted: true,
        })
      );
      this.stepper.selected!.completed = true;
      this.loading.set(false);
      this.router.navigate(['tutorials']);
      this.toast.success('Congrats on finishing the tutorial! Good Job!');
    } else if (this.progress?.completedAt) {
      this.router.navigate(['tutorials']);
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
