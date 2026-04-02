import { useState, useEffect } from "react"
import { jobApi, type Job } from "../api/jobs"
import { useJobEventStream } from "./usejobevenstream"
import { SafeStringifyJSON } from "../utils/json"
import { attempt } from "../utils/attempt"

export function useJobSelector(onSelectJobs?: (jobs: Job[]) => void) {
   const [selectedJobs, setSelectedJobs] = useState<Job[]>([])
   const [isCreating, setIsCreating] = useState(false)
   const [isProcessing, setIsProcessing] = useState<string | null>(null)
   const [isProcessingAll, setIsProcessingAll] = useState(false)

   const { status, connect, disconnect } = useJobEventStream({
      onJobUpdate: updatedJob => {
         setSelectedJobs(prev =>
            prev.map(job => {
               if (job.id !== updatedJob.id) return job

               return {
                  ...job,
                  status: updatedJob.status,
                  result: updatedJob.result ?? job.result,
                  error: updatedJob.error ?? job.error,
               }
            })
         )
      },
   })

   const handleCreateJob = async () => {
      setIsCreating(true)

      const payloadRes = SafeStringifyJSON({
         task: `Task ${Date.now()}`,
         data: Math.random() * 100,
      })
      if (payloadRes.isErr()) {
         setIsCreating(false)
         console.error("Failed to create job:", payloadRes.error)
         return
      }
      const payload = payloadRes.value

      const result = await jobApi.createJob(payload)
      if (result.success && result.job) {
         setSelectedJobs(prev => {
            const newJobs = [...prev, result.job!]
            onSelectJobs?.(newJobs)
            return newJobs
         })
      }
   }

   useEffect(() => {
      if (selectedJobs.length == 0) {
         return
      }
      connect(selectedJobs.map(j => j.id))
      return () => disconnect()
   }, [selectedJobs.length])

   const handleCreateRandomJobs = () => {
      const count = Math.floor(Math.random() * 3) + 2
      for (let i = 0; i < count; i++) {
         handleCreateJob()
      }
   }

   const handleProcessJob = async (jobId: string) => {
      setIsProcessing(jobId)

      const result = await attempt(jobApi.processJob(jobId))
      if (result.isErr()) {
         console.error("Failed to process job:", result.error)
         return
      }

      if (result.value.success) {
         setSelectedJobs(prev =>
            prev.map(job =>
               job.id === jobId
                  ? {
                       ...job,
                       status: "processing" as const,
                    }
                  : job
            )
         )
      }
   }

   const handleProcessAll = async () => {
      const idleJobs = selectedJobs.filter(job => job.status === "idle")
      if (idleJobs.length === 0) return

      setIsProcessingAll(true)

      await Promise.all(
         idleJobs.map(async job => {
            setIsProcessing(job.id)

            const result = await attempt(jobApi.processJob(job.id))
            if (result.isErr()) {
               console.error("Failed to process job:", result.error)
               return
            }

            if (result.value.success) {
               setSelectedJobs(prev =>
                  prev.map(j =>
                     j.id === job.id
                        ? {
                             ...j,
                             status: "processing" as const,
                          }
                        : j
                  )
               )
            }
         })
      )
      setIsProcessingAll(false)
   }

   const handleRemoveJob = (jobId: string) =>
      setSelectedJobs(prev => prev.filter(job => job.id !== jobId))

   return {
      isCreating,
      isProcessing,
      isProcessingAll,
      status,
      handleCreateJob,
      handleCreateRandomJobs,
      handleProcessJob,
      handleProcessAll,
      handleRemoveJob,
      selectedJobs,
      disconnect,
   }
}

export const formatDate = (dateStr: string) => {
   return new Intl.DateTimeFormat("en-US", {
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
   }).format(new Date(dateStr))
}
