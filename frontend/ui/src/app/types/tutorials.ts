export type Tutorial = 'branding' | 'agents' | 'registry'

export interface TutorialStepData {
  
}

export interface TutorialTaskData {
  tutorial: Tutorial;
  stepId: string;
  taskId: string;
  value: any;
}

/*
{

}
 */

export interface TutorialProgress {
  tutorial: Tutorial;
  steps: TutorialStep[];
  createdAt?: string;
}
