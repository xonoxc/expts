import { fromPromise, fromThrowable } from "neverthrow"

export function attempt<T, E = Error>(promise: Promise<T>) {
   return fromPromise(promise, err => err as E)
}

export function attemptSync<T, E = Error>(func: () => T) {
   return fromThrowable(func, err => err as E)()
}
