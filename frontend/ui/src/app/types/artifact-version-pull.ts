import {BaseArtifact, BaseArtifactVersion} from '../services/artifacts.service';

export interface ArtifactVersionPull {
  createdAt: string;
  remoteAddress?: string;
  userAccountName?: string;
  userAccountEmail?: string;
  customerOrganizationName?: string;
  artifact: BaseArtifact;
  artifactVersion: BaseArtifactVersion;
}
