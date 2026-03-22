import { err, fromPromise, ok, Result } from "neverthrow"
import { storage } from "./db"

export async function processJOB(jobId: string): Promise<Result<boolean, Error>> {
   const queingRes = await retryResult({
      fn: () =>
         storage.changeStatus(jobId, {
            status: "queued",
         }),
      retries: 3,
   })

   if (queingRes.isErr()) {
      return err(new Error("failed to change the status of the job"))
   }

   await sleep(3_000, jobId, "[QUEUED]started processing the job")

   const processingRes = await retryResult({
      fn: () =>
         storage.changeStatus(jobId, {
            status: "processing",
         }),
      retries: 3,
   })

   if (processingRes.isErr()) {
      return err(new Error("failed to change the status of the job"))
   }

   await sleep(3_000, jobId, "[PROCESSING]still processing the job ...")

   const completionRes = await retryResult({
      fn: () =>
         storage.changeStatus(jobId, {
            status: "completed",
            result: "Processing completed for the job",
         }),
      retries: 3,
   })

   if (completionRes.isErr()) {
      return err(new Error("failed to change the status of the job"))
   }

   console.log("job processing completed", jobId)

   return ok(true)
}

async function sleep(duration: number, jobId: string, message?: string) {
   await new Promise<void>(resolve => {
      setTimeout(() => {
         console.log(message ?? "awaiting io ...", jobId)
         resolve()
      }, duration)
   })
}

async function retryResult<T>({ fn, retries }: { fn: () => Promise<T>; retries: number }) {
   // 0 is passed as retry
   if (!retries) {
      throw new Error("Retries cannot be 0")
   }

   for (let i = 0; i < retries - 1; i++) {
      const result = await fromPromise(fn(), e => e as Error)
      if (result.isOk()) return result
   }

   return fromPromise(fn(), e => e as Error)
}
