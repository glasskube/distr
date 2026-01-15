import {UserAccount} from '@distr-sh/distr-sdk';
import {BaseArtifact, BaseArtifactVersion} from '../services/artifacts.service';

export interface ArtifactVersionPull {
  createdAt: string;
  remoteAddress?: string;
  userAccount?: UserAccount;
  artifact: BaseArtifact;
  artifactVersion: BaseArtifactVersion;
}
