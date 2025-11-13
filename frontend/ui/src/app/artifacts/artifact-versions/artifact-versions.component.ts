import {AsyncPipe} from '@angular/common';
import {Component, inject, resource} from '@angular/core';
import {ActivatedRoute, Router} from '@angular/router';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faBox, faDownload, faFile, faXmark} from '@fortawesome/free-solid-svg-icons';
import dayjs from 'dayjs';
import {
  catchError,
  combineLatest,
  distinctUntilChanged,
  filter,
  firstValueFrom,
  map,
  NEVER,
  Observable,
  startWith,
  Subject,
  switchMap,
  tap,
} from 'rxjs';
import {SemVer} from 'semver';
import {getRemoteEnvironment} from '../../../env/remote';
import {RelativeDatePipe} from '../../../util/dates';
import {BytesPipe} from '../../../util/units';
import {ClipComponent} from '../../components/clip.component';
import {UuidComponent} from '../../components/uuid';
import {
  ArtifactsService,
  ArtifactWithTags,
  HasDownloads,
  TaggedArtifactVersion,
} from '../../services/artifacts.service';
import {AuthService} from '../../services/auth.service';
import {OrganizationService} from '../../services/organization.service';
import {ArtifactsDownloadCountComponent, ArtifactsDownloadedByComponent, ArtifactsHashComponent} from '../components';
import {SecureImagePipe} from '../../../util/secureImage';
import {OverlayService} from '../../services/overlay.service';
import {RequireRoleDirective} from '../../directives/required-role.directive';
import {ToastService} from '../../services/toast.service';
import {getFormDisplayedError} from '../../../util/errors';

@Component({
  selector: 'app-artifact-tags',
  imports: [
    FaIconComponent,
    AsyncPipe,
    UuidComponent,
    RelativeDatePipe,
    ArtifactsDownloadCountComponent,
    ArtifactsDownloadedByComponent,
    ArtifactsHashComponent,
    ClipComponent,
    BytesPipe,
    SecureImagePipe,
    RequireRoleDirective,
  ],
  templateUrl: './artifact-versions.component.html',
})
export class ArtifactVersionsComponent {
  private readonly artifacts = inject(ArtifactsService);
  private readonly route = inject(ActivatedRoute);
  private readonly organization = inject(OrganizationService);
  private readonly overlay = inject(OverlayService);
  private readonly toast = inject(ToastService);

  protected readonly faBox = faBox;
  protected readonly faDownload = faDownload;
  protected readonly faFile = faFile;
  protected readonly faXmark = faXmark;

  private readonly refresh$ = new Subject<void>();

  protected readonly artifact$ = combineLatest([
    this.route.params.pipe(
      map((params) => params['id']?.trim()),
      distinctUntilChanged()
    ),
    this.refresh$.pipe(startWith(undefined)),
  ]).pipe(
    switchMap(([id]) => this.artifacts.getByIdAndCache(id)),
    map((artifact) => {
      if (artifact) {
        return {
          ...artifact,
          versions: (artifact.versions ?? []).map((v) => ({
            ...v,
            ...this.calcVersionDownloads(v),
          })),
        };
      }
      return undefined;
    })
  );

  protected readonly org = resource({
    loader: () => firstValueFrom(this.organization.get()),
  });
  private readonly remoteEnv = resource({
    loader: () => getRemoteEnvironment(),
  });

  public getArtifactUsage(artifact: ArtifactWithTags): string | undefined {
    if (!artifact.versions?.length) {
      // this should not actually happen
      return undefined;
    }
    const org = this.org.value();
    const env = this.remoteEnv.value();
    let url = `${org?.registryDomain ?? env?.registryHost ?? 'REGISTRY_DOMAIN'}/${org?.slug ?? 'ORG_SLUG'}/${artifact.name}`;
    const version = artifact.versions.find((it) => it.tags);
    if (!version) return;
    switch (version.inferredType) {
      case 'helm-chart':
        return `helm install <release-name> oci://${url} --version ${version.tags[0].name}`;
      case 'container-image':
        return `docker pull ${url}:${version.tags[0].name}`;
      default:
        return `oras pull ${url}:${version.tags[0].name}`;
    }
  }

  protected calcVersionDownloads(version: TaggedArtifactVersion): HasDownloads {
    let downloadsTotal = version.downloadsTotal ?? 0;
    let downloadedBySet: {[id: string]: boolean} = {};
    (version.downloadedByUsers ?? []).forEach((id: string) => (downloadedBySet[id] = true));
    for (let tag of version.tags) {
      (tag.downloads.downloadedByUsers ?? []).forEach((id: string) => (downloadedBySet[id] = true));
      downloadsTotal = downloadsTotal + (tag.downloads.downloadsTotal ?? 0);
    }
    const downloadedByUsers = Object.keys(downloadedBySet);
    return {
      downloadsTotal,
      downloadedByUsers,
      downloadedByCount: downloadedByUsers.length,
    };
  }

  public async uploadImage(data: ArtifactWithTags) {
    const fileId = await firstValueFrom(this.overlay.uploadImage({imageUrl: data.imageUrl}));
    if (!fileId || data.imageUrl?.includes(fileId)) {
      return;
    }
    await firstValueFrom(this.artifacts.patchImage(data.id!, fileId));
  }

  public deleteArtifactTag(artifact: ArtifactWithTags, version: TaggedArtifactVersion, tagName: string): void {
    this.overlay
      .confirm(
        `This will untag "${tagName}" from ${artifact.name}. The artifact version SHA (${version.digest.substring(0, 12)}) will remain in the database. Are you sure?`
      )
      .pipe(
        filter((result) => result === true),
        switchMap(() => this.artifacts.deleteArtifactTag(artifact, tagName)),
        catchError((e) => {
          const msg = getFormDisplayedError(e);
          if (msg) {
            this.toast.error(msg);
          }
          return NEVER;
        }),
        tap(() => {
          this.toast.success(`Tag "${tagName}" removed successfully`);
          this.refresh$.next();
        })
      )
      .subscribe();
  }
}
