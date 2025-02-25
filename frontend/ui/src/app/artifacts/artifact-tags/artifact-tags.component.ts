import {AsyncPipe} from '@angular/common';
import {Component, inject, resource} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faBox, faDownload} from '@fortawesome/free-solid-svg-icons';
import {firstValueFrom, switchMap} from 'rxjs';
import {RelativeDatePipe} from '../../../util/dates';
import {UuidComponent} from '../../components/uuid';
import {Artifact, ArtifactsService, ArtifactTag, ArtifactWithTags} from '../../services/artifacts.service';
import {ArtifactsVulnerabilityReportComponent} from '../artifacts-vulnerability-report.component';
import {ArtifactsDownloadCountComponent, ArtifactsDownloadedByComponent, ArtifactsHashComponent} from '../components';
import {OrganizationService} from '../../services/organization.service';
import {ClipComponent} from '../../components/clip.component';

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

  protected readonly faBox = faBox;
  protected readonly faDownload = faDownload;

  protected readonly artifact$ = this.route.params.pipe(switchMap((params) => this.artifacts.get(params['id'])));

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
}
