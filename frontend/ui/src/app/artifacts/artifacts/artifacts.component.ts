import {AsyncPipe} from '@angular/common';
import {Component, inject} from '@angular/core';
import {FormControl, FormGroup, ReactiveFormsModule} from '@angular/forms';
import {RouterLink} from '@angular/router';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faBox, faDownload, faMagnifyingGlass} from '@fortawesome/free-solid-svg-icons';
import {combineLatest, debounceTime, map, startWith} from 'rxjs';
import {UuidComponent} from '../../components/uuid';
import {ArtifactsService} from '../../services/artifacts.service';
import {ArtifactsDownloadCountComponent, ArtifactsDownloadedByComponent} from '../components';
import {faDocker} from '@fortawesome/free-brands-svg-icons';
import {AutotrimDirective} from '../../directives/autotrim.directive';

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
  ],
  templateUrl: './artifacts.component.html',
})
export class ArtifactsComponent {
  private readonly artifacts = inject(ArtifactsService);

  protected readonly faMagnifyingGlass = faMagnifyingGlass;
  protected readonly faBox = faBox;
  protected readonly faDownload = faDownload;

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
    ),
  );
  protected readonly faDocker = faDocker;
}
