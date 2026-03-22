import { useEffect, useRef, useState } from "react"
import { jobApi, type Job } from "../api/jobs"
import type { ConnectionStatus } from "../components/JobSelector/JobSelector"

interface UseJobEventStreamOptions {
   onJobUpdate?: (job: Job) => void
   onStatusChange?: (status: ConnectionStatus) => void
}

interface UseJobEventStreamReturn {
   status: ConnectionStatus
   connect: (jobIds: string[]) => void
   disconnect: () => void
}

export function useJobEventStream({
   onJobUpdate,
   onStatusChange,
}: UseJobEventStreamOptions = {}): UseJobEventStreamReturn {
   const [status, setStatus] = useState<ConnectionStatus>("disconnected")
   const eventSourceRef = useRef<EventSource | null>(null)

   const connect = (jobIds: string[]) => {
      if (eventSourceRef.current) {
         eventSourceRef.current.close()
      }

      setStatus("connecting")
      onStatusChange?.("connecting")

      const eventSource = jobApi.getJobStatus(jobIds)
      eventSourceRef.current = eventSource

      let isDone = false

      eventSource.onopen = () => {
         setStatus("connected")
         onStatusChange?.("connected")
      }

      eventSource.addEventListener("job-status", event => {
         try {
            const data = JSON.parse(event.data)

            console.log("Received job update:", data)

            if (data.id && data.status) {
               onJobUpdate?.(data)
            }
         } catch (err) {
            console.error("Parse error:", err)
         }
      })

      eventSource.addEventListener("done", () => {
         isDone = true
         setTimeout(() => {
            eventSource.close()
            eventSourceRef.current = null
         }, 3000)
      })

      eventSource.onerror = err => {
         if (isDone) return

         console.error("SSE error:", err)
         setStatus("disconnected")
         onStatusChange?.("disconnected")

         eventSourceRef.current = null
      }

      eventSource.onopen = () => {
         setStatus("connected")
         onStatusChange?.("connected")
      }
   }

   const disconnect = () => {
      if (eventSourceRef.current) {
         eventSourceRef.current.close()
         eventSourceRef.current = null
      }
      setStatus("disconnected")
      onStatusChange?.("disconnected")
   }

   useEffect(() => {
      return () => {
         eventSourceRef.current?.close()
      }
   }, [])

   return {
      status,
      connect,
      disconnect,
   }
}
