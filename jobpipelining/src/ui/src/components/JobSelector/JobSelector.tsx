import type { Job } from "../../api/jobs"
import { formatDate, useJobSelector } from "../../hooks/use-job-selector"
import "./JobSelector.css"

export type ConnectionStatus = "disconnected" | "connecting" | "connected"

interface JobSelectorProps {
   onSelectJobs?: (jobs: Job[]) => void
}

export function JobSelector({ onSelectJobs }: JobSelectorProps) {
   const {
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
   } = useJobSelector(onSelectJobs)

   return (
      <div className="min-h-screen bg-[#09090B] text-gray-200 font-sans p-8 md:p-12 relative overflow-hidden">
         <div className="absolute top-0 left-1/2 -translate-x-1/2 w-200 h-150 bg-[#7B61FF]/10 rounded-full blur-[120px] pointer-events-none -z-10 mix-blend-screen"></div>

         <div className="max-w-6xl mx-auto relative z-10">
            <div className="flex flex-col items-center text-center mb-20 pt-10">
               <div className="inline-flex items-center gap-3 px-4 py-1.5 rounded-full border border-white/10 bg-white/5 text-sm text-gray-300 mb-8 backdrop-blur-sm">
                  <span className="flex h-2 w-2 relative">
                     {status === "connected" && (
                        <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-[#7B61FF] opacity-75"></span>
                     )}
                     <span
                        className={`relative inline-flex rounded-full h-2 w-2 ${status === "connected" ? "bg-[#7B61FF]" : "bg-gray-500"}`}
                     ></span>
                  </span>
                  <span className="capitalize tracking-wide">{status}</span>
               </div>

               <h2 className="text-5xl md:text-6xl font-medium text-white tracking-tight mb-6">
                  Job Monitor
               </h2>
               <p className="text-gray-400 text-lg max-w-xl mx-auto mb-10">
                  Create, process, and track your background tasks in real-time.
               </p>

               <div className="flex flex-col sm:flex-row gap-4 justify-center w-full">
                  <button
                     onClick={handleCreateJob}
                     disabled={isCreating}
                     className="bg-[#7B61FF] text-white px-8 py-3 rounded-lg font-medium hover:bg-[#6A4BE0] transition-all disabled:opacity-50 min-w-45 shadow-[0_0_20px_rgba(123,97,255,0.3)] hover:shadow-[0_0_30px_rgba(123,97,255,0.5)]"
                  >
                     {isCreating ? "Creating..." : "Start creating"}
                  </button>

                  <button
                     onClick={handleCreateRandomJobs}
                     disabled={isCreating}
                     className="bg-transparent border border-white/10 text-white px-8 py-3 rounded-lg font-medium hover:bg-white/5 transition-colors disabled:opacity-50"
                  >
                     Explore random jobs
                  </button>

                  {selectedJobs.filter(j => j.status === "idle").length > 0 && (
                     <button
                        onClick={handleProcessAll}
                        disabled={isProcessingAll}
                        className="bg-emerald-600 text-white px-8 py-3 rounded-lg font-medium hover:bg-emerald-500 transition-all disabled:opacity-50 shadow-[0_0_20px_rgba(5,150,105,0.3)]"
                     >
                        {isProcessingAll
                           ? "Processing..."
                           : `Process all (${selectedJobs.filter(j => j.status === "idle").length})`}
                     </button>
                  )}
               </div>
            </div>

            {selectedJobs.length > 0 ? (
               <div className="space-y-8">
                  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                     {selectedJobs.map((job, index) => (
                        <div
                           key={job.id}
                           className="bg-[#121214] border border-white/10 rounded-xl flex flex-col transition-all hover:border-white/20 hover:shadow-2xl hover:shadow-[#7B61FF]/5"
                        >
                           {/* Card Header */}
                           <div className="p-5 border-b border-white/5 flex justify-between items-center bg-white/2 rounded-t-xl">
                              <div className="flex items-center gap-3">
                                 <div className="text-xs text-gray-500 uppercase tracking-wider">
                                    ID
                                 </div>
                                 <div className="font-mono text-sm text-gray-300">
                                    {job.id.slice(0, 8)}
                                 </div>
                              </div>
                              <StatusBadge
                                 status={job.status}
                                 selectedJobs={selectedJobs}
                                 index={index}
                                 jobId={job.id}
                              />
                           </div>

                           {/* Card Body */}
                           <div className="p-5 flex-1 flex flex-col gap-5 text-sm">
                              <div className="flex justify-between items-center">
                                 <span className="text-gray-500">Created</span>
                                 <span className="text-gray-300">{formatDate(job.createdAt)}</span>
                              </div>

                              <div className="space-y-2">
                                 <span className="text-gray-500 text-xs uppercase tracking-wider">
                                    Payload
                                 </span>
                                 <pre className="bg-black/50 p-3 text-xs text-gray-400 font-mono overflow-x-auto rounded-lg border border-white/5">
                                    {job.payload}
                                 </pre>
                              </div>

                              {job.result && (
                                 <div className="space-y-2">
                                    <span className="text-gray-500 text-xs uppercase tracking-wider">
                                       Result
                                    </span>
                                    <pre className="bg-black/50 p-3 text-xs text-[#7B61FF] font-mono overflow-x-auto rounded-lg border border-[#7B61FF]/20">
                                       {job.result}
                                    </pre>
                                 </div>
                              )}

                              {job.error && (
                                 <div className="space-y-2">
                                    <span className="text-rose-500/80 text-xs uppercase tracking-wider">
                                       Error
                                    </span>
                                    <pre className="bg-rose-950/30 p-3 text-xs text-rose-400 font-mono overflow-x-auto rounded-lg border border-rose-500/20">
                                       {job.error}
                                    </pre>
                                 </div>
                              )}
                           </div>

                           {/* Card Actions */}
                           <div className="p-5 flex gap-3 border-t border-white/5 mt-4 pt-4">
                              {job.status === "idle" && (
                                 <button
                                    onClick={() => handleProcessJob(job.id)}
                                    disabled={isProcessing === job.id}
                                    className="flex-1 bg-white/10 text-white py-2.5 rounded-lg text-sm font-medium hover:bg-white/20 transition-colors disabled:opacity-50"
                                 >
                                    {isProcessing === job.id ? "Processing..." : "Process"}
                                 </button>
                              )}
                              <button
                                 onClick={() => handleRemoveJob(job.id)}
                                 className="flex-none px-4 py-2.5 text-gray-500 text-sm font-medium hover:text-white transition-colors"
                              >
                                 Remove
                              </button>
                           </div>
                        </div>
                     ))}
                  </div>

                  <div className="flex justify-center pt-12">
                     <button
                        onClick={disconnect}
                        className="text-gray-500 hover:text-white px-8 py-3 text-sm font-medium transition-colors border-b border-transparent hover:border-white/20"
                     >
                        Stop Monitoring
                     </button>
                  </div>
               </div>
            ) : (
               <div className="mt-12"></div>
            )}
         </div>
      </div>
   )
}

const getStatusStyles = (status: Job["status"]) => {
   switch (status) {
      case "idle":
         return "bg-white/10 text-gray-300 border-white/5"
      case "queued":
         return "bg-blue-500/10 text-blue-400 border-blue-500/20"
      case "processing":
         return "bg-[#7B61FF]/10 text-[#7B61FF] border-[#7B61FF]/20"
      case "completed":
         return "bg-emerald-500/10 text-emerald-400 border-emerald-500/20"
      case "failed":
         return "bg-rose-500/10 text-rose-400 border-rose-500/20"
      default:
         return "bg-white/10 text-gray-300 border-white/5"
   }
}

function StatusBadge({
   status,
   selectedJobs,
   index,
   jobId,
}: {
   status: Job["status"]
   selectedJobs: Job[]
   index?: number
   jobId?: string
}) {
   if (index === selectedJobs.length - 1) {
      console.log("Last badge status", status)
      console.log("last job id:", jobId)
   }

   return (
      <span
         className={`text-[11px] px-2.5 py-1 uppercase tracking-wider font-semibold rounded-md border ${getStatusStyles(status)}`}
      >
         {status}
      </span>
   )
}
