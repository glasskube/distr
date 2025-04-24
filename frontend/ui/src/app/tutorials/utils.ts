import {TutorialProgress, TutorialProgressEvent} from '../types/tutorials';

export function getExistingTask(
  progress: TutorialProgress | undefined,
  stepId: string,
  taskId: string
): TutorialProgressEvent | undefined {
  return findTask(progress?.events ?? [], stepId, taskId);
}

export function getLastExistingTask(
  progress: TutorialProgress | undefined,
  stepId: string,
  taskId: string
): TutorialProgressEvent | undefined {
  return findTask((progress?.events ?? []).concat().reverse(), stepId, taskId);
}

function findTask(events: TutorialProgressEvent[], stepId: string, taskId: string): TutorialProgressEvent | undefined {
  return events.find((e) => e.stepId === stepId && e.taskId === taskId);
}
