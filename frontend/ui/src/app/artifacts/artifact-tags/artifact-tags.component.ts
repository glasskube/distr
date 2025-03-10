import {AsyncPipe} from '@angular/common';
import {Component, inject, resource} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faBox, faDownload, faFile, faLightbulb, faWarning} from '@fortawesome/free-solid-svg-icons';
import dayjs from 'dayjs';
import {firstValueFrom, map, Observable, switchMap} from 'rxjs';
import {SemVer} from 'semver';
import {RelativeDatePipe} from '../../../util/dates';
import {ClipComponent} from '../../components/clip.component';
import {UuidComponent} from '../../components/uuid';
import {ArtifactsService, ArtifactTag, ArtifactWithTags} from '../../services/artifacts.service';
import {AuthService} from '../../services/auth.service';
import {OrganizationService} from '../../services/organization.service';
import {ArtifactsVulnerabilityReportComponent} from '../artifacts-vulnerability-report.component';
import {ArtifactsDownloadCountComponent, ArtifactsDownloadedByComponent, ArtifactsHashComponent} from '../components';

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
  templateUrl: './artifact-tags.component.html',
})
export class ArtifactTagsComponent {
  private readonly artifacts = inject(ArtifactsService);
  private readonly route = inject(ActivatedRoute);
  private readonly organization = inject(OrganizationService);
  private readonly auth = inject(AuthService);

  protected readonly faBox = faBox;
  protected readonly faDownload = faDownload;
  protected readonly faLightbulb = faLightbulb;

  protected readonly artifact$ = this.route.params.pipe(switchMap((params) => this.artifacts.get(params['id'])));

  protected readonly updateTag$: Observable<ArtifactTag | null> = this.artifact$.pipe(
    map((artifact) => {
      const tagsSorted = [...artifact.tags]
        .sort((a, b) => {
          if (a.labels.some((l) => l.name === 'latest')) {
            return 1;
          }

          if (b.labels.some((l) => l.name === 'latest')) {
            return -1;
          }

          if (a.labels.length > 0 && b.labels.length > 0) {
            try {
              const aMax = a.labels
                .map((l) => new SemVer(l.name))
                .sort((a, b) => a.compare(b))
                .reverse()[0];
              const bMax = b.labels
                .map((l) => new SemVer(l.name))
                .sort((a, b) => a.compare(b))
                .reverse()[0];
              return aMax.compare(bMax);
            } catch (e) {
              console.warn(e);
              return dayjs(a.createdAt).diff(b.createdAt);
            }
          } else {
            return a.labels.length ? 1 : b.labels.length ? -1 : 0;
          }
        })
        .reverse();

      const newer = tagsSorted.slice(
        0,
        tagsSorted.findIndex((t) => t.downloadedByUsers.some((u) => u.id === this.auth.getClaims()?.sub))
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

  getOciUrl(
    artifact: ArtifactWithTags,
    tag: ArtifactTag | undefined = artifact.tags.find((it) => it.labels.some((l) => l.name === 'latest'))
  ) {
    const orgName = this.org.value()?.name?.replaceAll(/\W/g, '').toLowerCase();
    let url = `oci://${location.host}/${orgName ?? 'ORG_NAME'}/${artifact.name}`;
    if (tag) {
      return `${url}:${tag.labels[0].name}`;
    } else {
      return url;
    }
  }

  protected readonly faWarning = faWarning;
  protected readonly faFile = faFile;
}
