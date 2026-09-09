import {OverlayModule} from '@angular/cdk/overlay';
import {DatePipe} from '@angular/common';
import {Component, computed, inject, signal, TemplateRef} from '@angular/core';
import {rxResource} from '@angular/core/rxjs-interop';
import {FormControl, FormGroup, ReactiveFormsModule} from '@angular/forms';
import {
  AccessToken,
  AccessTokenSecret,
  AccessTokenSecretSlot,
  AccessTokenWithKey,
  CreateAccessTokenRequest,
  UserRole,
} from '@distr-sh/distr-sdk';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faClipboard, faKey, faPlus, faTrash, faTriangleExclamation, faXmark} from '@fortawesome/free-solid-svg-icons';
import dayjs from 'dayjs';
import {firstValueFrom} from 'rxjs';
import {isExpired, RelativeDatePipe} from '../../util/dates';
import {getFormDisplayedError} from '../../util/errors';
import {USER_ROLE_LABELS, UserRoleLabelPipe} from '../../util/user-role';
import {CreatedAccessTokenComponent} from '../components/created-access-token.component';
import {ExpiresAtPickerComponent} from '../components/expires-at-picker/expires-at-picker.component';
import {PageComponent} from '../components/page.component';
import {UserRoleSelectComponent} from '../components/user-role-select.component';
import {AutotrimDirective} from '../directives/autotrim.directive';
import {AccessTokensService} from '../services/access-tokens.service';
import {AuthService} from '../services/auth.service';
import {DialogRef, OverlayService} from '../services/overlay.service';
import {ToastService} from '../services/toast.service';

interface AccessTokenRow {
  token: AccessToken;
  expired: boolean;
  // A token without secrets predates them and is stored in plain text. Its first secret can only
  // be added at the cost of invalidating the token that is in circulation.
  legacy: boolean;
  secrets: AccessTokenSecret[];
  canCreateSecret: boolean;
  canDeleteSecret: boolean;
}

@Component({
  selector: 'app-access-tokens',
  imports: [
    ReactiveFormsModule,
    FaIconComponent,
    DatePipe,
    AutotrimDirective,
    OverlayModule,
    RelativeDatePipe,
    CreatedAccessTokenComponent,
    ExpiresAtPickerComponent,
    UserRoleSelectComponent,
    UserRoleLabelPipe,
    PageComponent,
  ],
  templateUrl: './access-tokens.component.html',
})
export class AccessTokensComponent {
  protected readonly faTrash = faTrash;
  protected readonly faPlus = faPlus;
  protected readonly faXmark = faXmark;
  protected readonly faClipboard = faClipboard;
  protected readonly faKey = faKey;
  protected readonly faTriangleExclamation = faTriangleExclamation;

  private readonly accessTokensService = inject(AccessTokensService);
  private readonly auth = inject(AuthService);
  private readonly toast = inject(ToastService);
  private readonly overlay = inject(OverlayService);

  private readonly accessTokens = rxResource({stream: () => this.accessTokensService.list()});

  protected readonly rows = computed<AccessTokenRow[]>(() =>
    (this.accessTokens.value() ?? []).map((token) => ({
      token,
      expired: isExpired(token),
      legacy: token.secrets.length === 0,
      secrets: token.secrets,
      canCreateSecret: token.secrets.length < 2,
      canDeleteSecret: token.secrets.length > 1,
    }))
  );

  protected drawer: DialogRef<void> | null = null;

  protected readonly currentUserRole = computed<UserRole | undefined>(() => this.auth.getClaims()?.role);
  protected readonly inheritOptionLabel = computed(() => {
    const role = this.currentUserRole();
    return role ? `Inherit (${USER_ROLE_LABELS[role]})` : 'Inherit from my role';
  });

  protected readonly editForm = new FormGroup({
    label: new FormControl('', {nonNullable: true}),
    expiresAt: new FormControl('', {nonNullable: true}),
    userRole: new FormControl<UserRole | undefined>(undefined),
  });

  protected readonly editFormLoading = signal(false);
  protected readonly createdToken = signal<AccessTokenWithKey | null>(null);

  public openDrawer(template: TemplateRef<unknown>) {
    this.hideDrawer();
    this.editForm.patchValue({
      label: '',
      expiresAt: dayjs()
        .add(dayjs.duration({days: 30}))
        .format('YYYY-MM-DD'),
      userRole: undefined,
    });
    this.drawer = this.overlay.showDrawer(template);
  }

  public hideDrawer() {
    this.drawer?.dismiss();
  }

  public async createAccessToken() {
    this.editFormLoading.set(true);
    const request: CreateAccessTokenRequest = {};
    if (this.editForm.value.label) {
      request.label = this.editForm.value.label;
    }
    if (this.editForm.value.expiresAt) {
      request.expiresAt = new Date(this.editForm.value.expiresAt);
    }
    if (this.editForm.value.userRole) {
      request.userRole = this.editForm.value.userRole;
    }
    try {
      this.createdToken.set(await firstValueFrom(this.accessTokensService.create(request)));
      this.toast.success('token created');
      this.hideDrawer();
      this.accessTokens.reload();
    } finally {
      this.editFormLoading.set(false);
    }
  }

  public async deleteAccessToken(row: AccessTokenRow) {
    if (await firstValueFrom(this.overlay.confirm(`Really delete token '${row.token.label}'?`))) {
      try {
        await firstValueFrom(this.accessTokensService.delete(row.token.id!));
        this.accessTokens.reload();
      } catch (e) {
        this.showError(e);
      }
    }
  }

  public async createSecret(row: AccessTokenRow) {
    const confirmation = row.legacy
      ? `Token '${row.token.label}' is still stored in plain text. Securing it replaces it with a new token, ` +
        'so the one currently in use stops working. Continue?'
      : `Add a second secret to token '${row.token.label}'?`;
    if (!(await firstValueFrom(this.overlay.confirm(confirmation)))) {
      return;
    }
    try {
      this.createdToken.set(await firstValueFrom(this.accessTokensService.createSecret(row.token.id!)));
      this.toast.success('secret created');
      this.accessTokens.reload();
    } catch (e) {
      this.showError(e);
    }
  }

  public async deleteSecret(row: AccessTokenRow, slot: AccessTokenSecretSlot) {
    const confirmation =
      `Really delete secret ${slot} of token '${row.token.label}'? ` +
      'Everything that still authenticates with it stops working.';
    if (await firstValueFrom(this.overlay.confirm(confirmation))) {
      try {
        await firstValueFrom(this.accessTokensService.deleteSecret(row.token.id!, slot));
        this.accessTokens.reload();
      } catch (e) {
        this.showError(e);
      }
    }
  }

  private showError(e: unknown) {
    const message = getFormDisplayedError(e);
    if (message) {
      this.toast.error(message);
    }
  }
}
