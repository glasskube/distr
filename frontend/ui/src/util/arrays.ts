export type Predicate<T, R> = (arg: T) => R;

export function distinctBy<T>(predicate: Predicate<T, unknown>): (input: T[]) => T[] {
  return (input) =>
    input.filter((value: T, index, self) => {
      return self.findIndex((element) => predicate(element) === predicate(value)) === index;
    });
}

export function compareBy<T>(predicate: Predicate<T, string>): (a: T, b: T) => number {
  return (a, b) => predicate(a).localeCompare(predicate(b));
}

export function maxBy<T, E>(
  input: T[],
  predicate: Predicate<T, E>,
  cmp: (a: E, b: E) => boolean = (a, b) => a > b
): T | undefined {
  let max: T | undefined;
  let maxp: E | undefined;
  for (const el of input) {
    const elp = predicate(el);
    if (maxp === undefined || cmp(elp, maxp)) {
      max = el;
      maxp = elp;
    }
  }
  return max;
}
