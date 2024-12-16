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
