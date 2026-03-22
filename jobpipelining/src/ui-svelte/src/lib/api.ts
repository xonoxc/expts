const API_BASE = "http://localhost:3000"

export interface Job {
   id: string
   status: "idle" | "queued" | "processing" | "completed" | "failed"
   payload: string
   result?: string
   error?: string
   createdAt: string
   updatedAt: string
}

interface ApiResponse<T> {
   success: boolean
   message: string
   job?: T
}

export const jobApi = {
   async createJob(payload: string): Promise<ApiResponse<Job>> {
      const job: Job = {
         id: crypto.randomUUID(),
         status: "idle",
         payload,
         createdAt: new Date().toISOString(),
         updatedAt: new Date().toISOString(),
      }

      const res = await fetch(`${API_BASE}/api/create-job`, {
         method: "POST",
         headers: { "Content-Type": "application/json" },
         body: JSON.stringify(job),
      })

      if (!res.ok) {
         throw new Error(`Failed to create job: ${res.statusText}`)
      }

      return res.json()
   },

   async processJob(jobId: string): Promise<ApiResponse<void>> {
      const res = await fetch(`${API_BASE}/api/process/${jobId}`, {
         method: "POST",
      })

      if (!res.ok) {
         throw new Error(`Failed to process job: ${res.statusText}`)
      }

      return res.json()
   },

   getJobStatus(jobIds: string[]): EventSource {
      return new EventSource(`${API_BASE}/api/jobs/status?jobIds=${jobIds.join(",")}`)
   },
}
