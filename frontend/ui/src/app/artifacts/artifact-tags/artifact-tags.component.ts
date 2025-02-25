import {AsyncPipe, DatePipe} from '@angular/common';
import {Component, inject} from '@angular/core';
import {ActivatedRoute} from '@angular/router';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faBox, faDownload} from '@fortawesome/free-solid-svg-icons';
import {switchMap} from 'rxjs';
import {UuidComponent} from '../../components/uuid';
import {ArtifactsService} from '../../services/artifacts.service';
import {RelativeDatePipe} from '../../../util/dates';
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
  ],
  templateUrl: './artifact-tags.component.html',
})
export class ArtifactTagsComponent {
  private readonly artifacts = inject(ArtifactsService);
  private readonly route = inject(ActivatedRoute);

  protected readonly faBox = faBox;
  protected readonly faDownload = faDownload;

  protected readonly artifact$ = this.route.params.pipe(switchMap((params) => this.artifacts.get(params['id'])));
}
