import {AsyncPipe} from '@angular/common';
import {Component, inject, resource} from '@angular/core';
import {ActivatedRoute, Router} from '@angular/router';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faBox, faDownload, faFile, faTrash} from '@fortawesome/free-solid-svg-icons';
import dayjs from 'dayjs';
import {catchError, distinctUntilChanged, filter, firstValueFrom, map, NEVER, Observable, Subject, switchMap, tap} from 'rxjs';
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
import {ArtifactsVulnerabilityReportComponent} from '../artifacts-vulnerability-report.component';
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
    ArtifactsVulnerabilityReportComponent,
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
  private readonly auth = inject(AuthService);
  private readonly overlay = inject(OverlayService);
  private readonly toast = inject(ToastService);
  private readonly router = inject(Router);

  protected readonly faBox = faBox;
  protected readonly faDownload = faDownload;
  protected readonly faFile = faFile;
  protected readonly faTrash = faTrash;

  private readonly refresh$ = new Subject<void>();

  protected readonly artifact$ = this.route.params.pipe(
    map((params) => params['id']?.trim()),
    distinctUntilChanged(),
    switchMap((id: string) => this.artifacts.getByIdAndCache(id)),
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

  protected readonly updateTag$: Observable<TaggedArtifactVersion | null> = this.artifact$.pipe(
    filter((a) => !!a),
    map((artifact) => {
      const tagsSorted = [...artifact.versions]
        .sort((a, b) => {
          if (a.tags.some((l) => l.name === 'latest')) {
            return 1;
          }

          if (b.tags.some((l) => l.name === 'latest')) {
            return -1;
          }

          if (a.tags.length > 0 && b.tags.length > 0) {
            try {
              const aMax = a.tags
                .map((l) => new SemVer(l.name))
                .sort((a, b) => a.compare(b))
                .reverse()[0];
              const bMax = b.tags
                .map((l) => new SemVer(l.name))
                .sort((a, b) => a.compare(b))
                .reverse()[0];
              return aMax.compare(bMax);
            } catch (e) {
              console.warn(e);
              return dayjs(a.createdAt).diff(b.createdAt);
            }
          } else {
            return a.tags.length ? 1 : b.tags.length ? -1 : 0;
          }
        })
        .reverse();

      const newer = tagsSorted.slice(
        0,
        tagsSorted.findIndex((t) => (t.downloadedByUsers ?? []).some((u) => u === this.auth.getClaims()?.sub))
      );

      if (newer.length > 0) {
        return newer[0];
      }

      return null;
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
    const version = artifact.versions[0];
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

  public deleteArtifactVersion(artifact: ArtifactWithTags, version: TaggedArtifactVersion): void {
    const tagNames = version.tags.map((t) => t.name).join(', ');
    this.overlay
      .confirm({
        message: {
          message: `This will permanently delete version ${tagNames} (${version.digest.substring(0, 12)}) from ${artifact.name}. Users will no longer be able to download this specific version. Are you sure?`,
        },
      })
      .pipe(
        filter((result) => result === true),
        switchMap(() => this.artifacts.deleteArtifactVersion(artifact.id, version.id)),
        catchError((e) => {
          const msg = getFormDisplayedError(e);
          if (msg) {
            this.toast.error(msg);
          }
          return NEVER;
        }),
        tap(() => {
          this.toast.success('Artifact version deleted successfully');
          this.refresh$.next();
        })
      )
      .subscribe();
  }
}
