import {AsyncPipe, DatePipe, JsonPipe} from '@angular/common';
import {Component, inject, TemplateRef} from '@angular/core';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faMagnifyingGlass, faPlus, faTrash, faXmark} from '@fortawesome/free-solid-svg-icons';
import {firstValueFrom, startWith, Subject, switchMap} from 'rxjs';
import {AccessTokensService} from '../services/access-tokens.service';
import {DialogRef, OverlayService} from '../services/overlay.service';
import {AccessToken, AccessTokenWithKey, CreateAccessTokenRequest} from '../types/access-token';
import {RequireRoleDirective} from '../directives/required-role.directive';
import {drawerFlyInOut} from '../animations/drawer';
import {AutotrimDirective} from '../directives/autotrim.directive';
import {OverlayModule} from '@angular/cdk/overlay';
import dayjs, {duration} from 'dayjs';
import {RelativeDatePipe} from '../../util/dates';
import {ToastService} from '../services/toast.service';

@Component({
  selector: 'app-access-tokens',
  imports: [
    ReactiveFormsModule,
    FaIconComponent,
    AsyncPipe,
    DatePipe,
    RequireRoleDirective,
    AutotrimDirective,
    OverlayModule,
    RelativeDatePipe,
    JsonPipe,
  ],
  templateUrl: './access-tokens.component.html',
  animations: [drawerFlyInOut],
})
export class AccessTokensComponent {
  protected readonly faMagnifyingGlass = faMagnifyingGlass;
  protected readonly faTrash = faTrash;
  protected readonly faPlus = faPlus;
  protected readonly faXmark = faXmark;

  private readonly accessTokens = inject(AccessTokensService);
  private readonly refresh$ = new Subject<void>();
  protected readonly accessTokens$ = this.refresh$.pipe(
    startWith(0),
    switchMap(() => this.accessTokens.list())
  );

  private readonly toast = inject(ToastService);

  private readonly overlay = inject(OverlayService);
  protected drawer: DialogRef<void> | null = null;

  protected readonly editForm = new FormGroup({
    label: new FormControl('', {nonNullable: true}),
    expiresAt: new FormControl('', {nonNullable: true}),
  });

  protected editFormLoading = false;
  protected createdToken: AccessTokenWithKey | null = null;

  public isExpired(token: AccessToken): boolean {
    return !token.expiresAt || dayjs(token.expiresAt).isBefore();
  }

  public openDrawer(template: TemplateRef<unknown>) {
    this.hideDrawer();
    this.editForm.patchValue({
      label: '',
      expiresAt: dayjs()
        .add(dayjs.duration({days: 30}))
        .format('YYYY-MM-DD'),
    });
    this.drawer = this.overlay.showDrawer(template);
  }

  public hideDrawer() {
    this.drawer?.dismiss();
  }

  public async createAccessToken() {
    this.editFormLoading = true;
    const request: CreateAccessTokenRequest = {};
    if (this.editForm.value.label) {
      request.label = this.editForm.value.label;
    }
    if (this.editForm.value.expiresAt) {
      request.expiresAt = new Date(this.editForm.value.expiresAt);
    }
    try {
      this.createdToken = await firstValueFrom(this.accessTokens.create(request));
      this.toast.success('token created');
      this.hideDrawer();
      this.refresh$.next();
    } finally {
      this.editFormLoading = false;
    }
  }

  public async deleteAccessToken(accessToken: AccessToken) {
    if (await firstValueFrom(this.overlay.confirm(`Really delete token '${accessToken.label}'`))) {
      try {
        await firstValueFrom(this.accessTokens.delete(accessToken.id!));
        this.refresh$.next();
      } catch (e) {}
    }
  }
}
