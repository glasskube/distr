import {AsyncPipe} from '@angular/common';
import {Component, inject} from '@angular/core';
import {FormControl, FormGroup, ReactiveFormsModule} from '@angular/forms';
import {RouterLink} from '@angular/router';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faBox, faLightbulb, faMagnifyingGlass, faTrash, faUserCircle} from '@fortawesome/free-solid-svg-icons';
import {catchError, combineLatest, filter, map, NEVER, startWith, switchMap, tap} from 'rxjs';
import {UuidComponent} from '../../components/uuid';
import {ArtifactsService, ArtifactWithTags} from '../../services/artifacts.service';
import {ArtifactsDownloadCountComponent, ArtifactsDownloadedByComponent} from '../components';
import {AutotrimDirective} from '../../directives/autotrim.directive';
import {getRemoteEnvironment} from '../../../env/remote';
import {fromPromise} from 'rxjs/internal/observable/innerFrom';
import {OrganizationService} from '../../services/organization.service';
import {RequireRoleDirective} from '../../directives/required-role.directive';
import {SecureImagePipe} from '../../../util/secureImage';
import {OverlayService} from '../../services/overlay.service';
import {ToastService} from '../../services/toast.service';
import {getFormDisplayedError} from '../../../util/errors';

@Component({
  selector: 'app-artifacts',
  imports: [
    ReactiveFormsModule,
    AsyncPipe,
    FaIconComponent,
    UuidComponent,
    RouterLink,
    ArtifactsDownloadCountComponent,
    ArtifactsDownloadedByComponent,
    AutotrimDirective,
    RequireRoleDirective,
    SecureImagePipe,
  ],
  templateUrl: './artifacts.component.html',
})
export class ArtifactsComponent {
  private readonly artifacts = inject(ArtifactsService);
  private readonly overlay = inject(OverlayService);
  private readonly toast = inject(ToastService);

  protected readonly faMagnifyingGlass = faMagnifyingGlass;
  protected readonly faBox = faBox;
  protected readonly faTrash = faTrash;

  protected readonly filterForm = new FormGroup({
    search: new FormControl(''),
  });

  protected readonly artifacts$ = this.artifacts.list();

  protected readonly filteredArtifacts$ = combineLatest([
    this.artifacts$,
    this.filterForm.valueChanges.pipe(startWith(this.filterForm.value)),
  ]).pipe(
    map(([artifacts, formValue]) =>
      artifacts.filter((it) => !formValue.search || it.name.toLowerCase().includes(formValue.search.toLowerCase()))
    )
  );
  protected readonly faLightbulb = faLightbulb;

  private readonly organizationService = inject(OrganizationService);
  protected readonly registrySlug$ = this.organizationService.get().pipe(map((org) => org.slug));
  protected readonly registryHost$ = combineLatest([
    fromPromise(getRemoteEnvironment()),
    this.organizationService.get(),
  ]).pipe(map(([env, org]) => org.registryDomain ?? env.registryHost));
  protected readonly faUserCircle = faUserCircle;

  public deleteArtifact(artifact: ArtifactWithTags): void {
    this.overlay
      .confirm(
        `This will permanently delete ${artifact.name} and all its versions. Users will no longer be able to download this artifact. Are you sure?`
      )
      .pipe(
        filter((result) => result === true),
        switchMap(() => this.artifacts.deleteArtifact(artifact.id)),
        catchError((e) => {
          const msg = getFormDisplayedError(e);
          if (msg) {
            this.toast.error(msg);
          }
          return NEVER;
        }),
        tap(() => {
          this.toast.success('Artifact deleted successfully');
        })
      )
      .subscribe();
  }
}
