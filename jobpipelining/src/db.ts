import { drizzle, BunSQLiteDatabase } from "drizzle-orm/bun-sqlite"
import { eq } from "drizzle-orm"
import { jobs, type Job } from "./db.schema"
import { attempt } from "./ui/src/utils/attempt"

export type JobState =
   | { status: "idle" }
   | { status: "queued" }
   | { status: "processing" }
   | {
        status: "completed"
        result: string
     }
   | {
        status: "failed"
        error: string
     }

class Storage {
   #db: BunSQLiteDatabase<{ jobs: typeof jobs }>

   constructor() {
      this.#db = drizzle("db.sqlite3", {
         schema: {
            jobs,
         },
      })
   }

   public getInstance() {
      return this.#db
   }

   public async checkJobStatus(jobId: string) {
      return attempt(this.#db.select().from(jobs).where(eq(jobs.id, jobId)))
   }

   public async getJob(jobId: string) {
      return attempt(this.#db.select().from(jobs).where(eq(jobs.id, jobId)))
   }

   public async createJOB(jobData: Job) {
      return attempt(
         this.#db
            .insert(jobs)
            .values({ ...jobData })
            .returning()
      )
   }

   public async changeStatus(jobId: string, jobState: JobState) {
      const updateValues: Record<string, string> = {
         status: jobState.status,
      }

      if (jobState.status === "completed") {
         updateValues.result = jobState.result
      } else if (jobState.status === "failed") {
         updateValues.error = jobState.error
      }

      updateValues.updatedAt = new Date().toISOString()

      const current = await this.#db.select().from(jobs).where(eq(jobs.id, jobId))

      const currentStatus = current[0]?.status

      const order = ["idle", "queued", "processing", "completed"]

      if (currentStatus && order.indexOf(jobState.status) < order.indexOf(currentStatus)) {
         return
      }

      return attempt(this.#db.update(jobs).set(updateValues).where(eq(jobs.id, jobId)).returning())
   }
}

export const storage = new Storage()

export type AppStore = typeof storage
