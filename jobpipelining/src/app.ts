import { Hono } from "hono"
import { cors } from "hono/cors"
import { storage } from "./db"

import type { Job } from "./db.schema"
import { processJOB } from "./processor"
import { betterLogger } from "./logger"
import { streamSSE } from "hono/streaming"

const app = new Hono()

app.use(
   cors({
      origin: "*",
      allowMethods: ["GET", "POST", "OPTIONS"],
      allowHeaders: ["Content-Type"],
   })
)

app.use(betterLogger())

app.get("/api", c => {
   return c.json({
      success: true,
      message: "hello form api",
   })
})

app.post("/api/create-job", async ctx => {
   const body = await ctx.req.json<Job>()

   const result = await storage.createJOB(body)
   if (result.isErr()) {
      console.error("failed to create job:", result.error)
      return ctx.json(
         {
            success: false,
            message: "something went wrong cannot create jobs",
         },
         500
      )
   }

   const [job] = result.value

   return ctx.json(
      {
         success: true,
         message: "job created successfully!",
         job,
      },
      201
   )
})

app.post("/api/process/:jobId", async ctx => {
   const jobId = ctx.req.param("jobId")

   const result = await storage.getJob(jobId)
   if (result.isErr()) {
      console.error("failed to get job:", result.error)
      return ctx.json(
         {
            success: false,
            message: "failed to fetch job",
         },
         500
      )
   }

   if (!result.value.length) {
      return ctx.json(
         {
            success: false,
            message: "job not found",
         },
         404
      )
   }

   const procRes = await processJOB(jobId)
   if (procRes.isErr()) {
      console.error("error while processing job:", procRes.error)
      return ctx.json(
         {
            success: false,
            message: "failure processing job",
         },
         500
      )
   }

   return ctx.json(
      {
         success: true,
         message: "processed successfully",
      },
      201
   )
})

app.get("/api/jobs/status", async ctx => {
   const jobIds = ctx.req.query("jobIds")?.split(",").filter(Boolean) ?? []

   const completedJobs = new Set<string>()
   const lastStatus = new Map<string, string>()

   return streamSSE(ctx, async stream => {
      await stream.writeSSE({
         event: "connected",
         data: JSON.stringify({ message: "connected" }),
      })

      while (true) {
         await stream.writeSSE({
            event: "ping",
            data: "keep-alive",
         })

         if (completedJobs.size === jobIds.length) {
            await stream.sleep(5000)
            await stream.writeSSE({
               event: "done",
               data: "all jobs completed",
            })
            return
         }

         for (const jobId of jobIds) {
            if (completedJobs.has(jobId)) continue

            const status = await storage.checkJobStatus(jobId)
            if (status.isErr()) {
               completedJobs.add(jobId)
               continue
            }

            const job = status.value[0]
            if (!job) {
               completedJobs.add(jobId)
               continue
            }

            if (job.status !== lastStatus.get(jobId)) {
               lastStatus.set(jobId, job.status as string)

               await stream.writeSSE({
                  event: "job-status",
                  data: JSON.stringify({
                     id: job.id,
                     status: job.status,
                     result: job.result ?? null,
                     error: job.error ?? null,
                  }),
               })
            }

            if (job.status === "completed" || job.status === "failed") {
               completedJobs.add(jobId)
            }
         }
         await stream.sleep(1000)
      }
   })
})

export { app }
