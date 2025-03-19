import {AsyncPipe} from '@angular/common';
import {Component, inject, resource} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faBox, faDownload, faFile, faLightbulb, faWarning} from '@fortawesome/free-solid-svg-icons';
import dayjs from 'dayjs';
import {
  combineLatestWith,
  distinctUntilChanged,
  filter,
  firstValueFrom,
  map,
  Observable,
  of,
  switchMap,
  tap,
  withLatestFrom,
} from 'rxjs';
import {SemVer} from 'semver';
import {RelativeDatePipe} from '../../../util/dates';
import {ClipComponent} from '../../components/clip.component';
import {UuidComponent} from '../../components/uuid';
import {ArtifactsService, TaggedArtifactVersion, ArtifactWithTags} from '../../services/artifacts.service';
import {AuthService} from '../../services/auth.service';
import {OrganizationService} from '../../services/organization.service';
import {ArtifactsVulnerabilityReportComponent} from '../artifacts-vulnerability-report.component';
import {ArtifactsDownloadCountComponent, ArtifactsDownloadedByComponent, ArtifactsHashComponent} from '../components';
import {getRemoteEnvironment} from '../../../env/remote';

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
  ],
  templateUrl: './artifact-versions.component.html',
})
export class ArtifactVersionsComponent {
  private readonly artifacts = inject(ArtifactsService);
  private readonly route = inject(ActivatedRoute);
  private readonly organization = inject(OrganizationService);
  private readonly auth = inject(AuthService);

  protected readonly faBox = faBox;
  protected readonly faDownload = faDownload;
  protected readonly faLightbulb = faLightbulb;

  protected readonly artifact$ = this.route.params.pipe(
    map((params) => params['id']?.trim()),
    distinctUntilChanged(),
    switchMap((id: string) => this.artifacts.getByIdAndCache(id))
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

  getOciUrl(
    artifact: ArtifactWithTags,
    tag: TaggedArtifactVersion | undefined = artifact.versions.find((it) => it.tags.some((l) => l.name === 'latest'))
  ) {
    const orgName = this.org.value()?.slug;
    let url = `oci://${this.remoteEnv.value()?.artifactsHost}/${orgName ?? 'ORG_SLUG'}/${artifact.name}`;
    if (tag) {
      return `${url}:${tag.tags[0].name}`;
    } else {
      return url;
    }
  }

  protected readonly faWarning = faWarning;
  protected readonly faFile = faFile;
}
