import {TutorialProgress, TutorialProgressEvent} from '../types/tutorials';

export function getExistingTask(
  progress: TutorialProgress | undefined,
  stepId: string,
  taskId: string
): TutorialProgressEvent | undefined {
  return (progress?.events ?? []).find((e) => e.stepId === stepId && e.taskId === taskId);
}
