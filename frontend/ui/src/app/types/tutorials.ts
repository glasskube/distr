export type Tutorial = 'branding' | 'agents' | 'registry'

export interface TutorialStepData {

}

export interface TutorialProgressEvent {
  stepId: string;
  taskId: string;
  value?: any;
}

export interface TutorialProgressRequest extends TutorialProgressEvent {
  markCompleted?: boolean;
}

/*
{

}
 */

export interface TutorialTaskData {
  [key: string]: {
    value?: boolean | string | number
  };
}

export interface TutorialProgressData {
  [key: string]: TutorialTaskData;
}

export interface TutorialProgress {
  tutorial: Tutorial;
  // steps: TutorialStep[];
  createdAt?: string;
  completedAt?: string;
  // data?: TutorialProgressData;
  events: TutorialProgressEvent[];
}
