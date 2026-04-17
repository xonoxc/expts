import { attempt } from "../utils/attempt"
import { SafeParseJSON, SafeStringifyJSON } from "../utils/json"

import { jobApi, type Job } from "./api"

export function createJobStore() {
   let jobs = $state<Job[]>([])
   let isCreating = $state(false)
   let isProcessing = $state<string | null>(null)
   let isProcessingAll = $state(false)
   let connectionStatus = $state<"disconnected" | "connecting" | "connected">("disconnected")

   let eventSource: EventSource | null = $state(null)
   let currentJobIds = $state("")

   let idleJobsCount = $derived(jobs.filter(j => j.status === "idle").length)

   function connect(jobIds: string[]) {
      const newIds = jobIds.join(",")

      if (currentJobIds === newIds) return

      currentJobIds = newIds

      if (eventSource) {
         eventSource.close()
         eventSource = null
      }

      connectionStatus = "connecting"

      eventSource = jobApi.getJobStatus(jobIds)

      eventSource.onopen = () => {
         connectionStatus = "connected"
         console.log("Svelte SSE: Connected")
      }

      eventSource.addEventListener("job-status", event => {
         const dataRes = SafeParseJSON(event.data)
         if (dataRes.isErr()) {
            console.error("Svelte SSE: Parse error", dataRes.error)
            return
         }
         const data = dataRes.value

         if (data.id && data.status) {
            jobs = jobs.map(job =>
               job.id === data.id
                  ? {
                       ...job,
                       ...data,
                    }
                  : job
            )
         }
      })

      eventSource.addEventListener("done", () => {
         console.log("Svelte SSE: Done")
         connectionStatus = "disconnected"
         currentJobIds = ""
         if (eventSource) {
            eventSource.close()
            eventSource = null
         }
      })

      eventSource.onerror = () => {
         console.error("Svelte SSE: Error")
         connectionStatus = "disconnected"
         currentJobIds = ""
         if (eventSource) {
            eventSource.close()
            eventSource = null
         }
      }
   }

   function disconnect() {
      if (eventSource) {
         eventSource.close()
         eventSource = null
      }
      currentJobIds = ""
      connectionStatus = "disconnected"
   }

   $effect(() => {
      const activeJobs = jobs.filter(
         j => j.status === "idle" || j.status === "processing" || j.status === "queued"
      )
      if (activeJobs.length === 0) return
      connect(activeJobs.map(j => j.id))
   })

	async function handleCreateJob() {
		isCreating = true

		const jsonPayload= {
			task: `Task ${Date.now()}`,
			data: Math.random() * 100,
		}

		const finalRes = 
		await SafeStringifyJSON(jsonPayload)
		.asyncAndThen(
			payload => attempt(jobApi.createJob(payload)),
		)

		finalRes.match(
			res=> {
				if (res.success && res.job) {
					jobs = [...jobs, res.job]
				}
			},
			err=> {
				console.error("Failed to create job:", err)
			}
		)

		isCreating = false
	}

   async function handleCreateRandomJobs() {
      const count = Math.floor(Math.random() * 3) + 2
      for (let i = 0; i < count; i++) {
         await handleCreateJob()
      }
   }

   async function handleProcessJob(jobId: string) {
      isProcessing = jobId

      const resultRes = await attempt(jobApi.processJob(jobId))
      if (resultRes.isErr()) {
         isProcessing = null
         console.error("Failed to process job:", resultRes.error)
         return
      }

      const result = resultRes.value

      if (result.success) {
         jobs = jobs.map(job => (job.id === jobId ? {
				...job, status: "processing" 
			} : job))
      }
   }

   async function handleProcessAll() {
      const idleJobs = jobs.filter(job => job.status === "idle")
      if (idleJobs.length === 0) return

      isProcessingAll = true

      await Promise.all(
         idleJobs.map(async job => {
            isProcessing = job.id

            const resultRes = await attempt(jobApi.processJob(job.id))
            if (resultRes.isErr()) {
               console.error("Failed to process job:", resultRes.error)
               return
            }

            const result = resultRes.value

            if (result.success) {
               jobs = jobs.map(j => (j.id === job.id ? { ...j, status: "processing" } : j))
            }
         })
      )

      isProcessingAll = false
      isProcessing = null
   }

   function handleRemoveJob(jobId: string) {
      jobs = jobs.filter(job => job.id !== jobId)
   }

   function getConnectionDotColor(status: string): string {
      switch (status) {
         case "connected":
            return "bg-purple-500"
         case "connecting":
            return "bg-yellow-500"
         default:
            return "bg-gray-500"
      }
   }

   return {
      getConnectionDotColor,
      handleCreateJob,
      handleCreateRandomJobs,
      handleRemoveJob,

      jobs,
      isCreating,
      isProcessing,
      isProcessingAll,
      connectionStatus,
      idleJobsCount,

      handleProcessJob,
      handleProcessAll,
      disconnect,
   }
}
