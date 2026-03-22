<script lang="ts">
   import type { Job } from "./api"

   interface Props {
      job: Job
      onProcess: (id: string) => void
      onRemove: (id: string) => void
      isProcessing: boolean
   }

   let { job, onProcess, onRemove, isProcessing }: Props = $props()

   function getStatusStyles(status: Job["status"]): string {
      switch (status) {
         case "idle":
            return "bg-white/10 text-gray-300 border-white/5"
         case "queued":
            return "bg-blue-500/10 text-blue-400 border-blue-500/20"
         case "processing":
            return "bg-purple-500/10 text-purple-400 border-purple-500/20"
         case "completed":
            return "bg-emerald-500/10 text-emerald-400 border-emerald-500/20"
         case "failed":
            return "bg-rose-500/10 text-rose-400 border-rose-500/20"
         default:
            return "bg-white/10 text-gray-300 border-white/5"
      }
   }

   function formatDate(dateStr: string): string {
      return new Intl.DateTimeFormat("en-US", {
         month: "short",
         day: "numeric",
         hour: "2-digit",
         minute: "2-digit",
      }).format(new Date(dateStr))
   }
</script>

<div
   class="bg-zinc-900 border border-white/10 rounded-xl flex flex-col transition-all hover:border-white/20 hover:shadow-2xl hover:shadow-purple-500/5"
>
   <!-- Card Header -->
   <div class="p-5 border-b border-white/5 flex justify-between items-center bg-white/2 rounded-t-xl">
      <div class="flex items-center gap-3">
         <div class="text-xs text-gray-500 uppercase tracking-wider">ID</div>
         <div class="font-mono text-sm text-gray-300">
            {job.id.slice(0, 8)}
         </div>
      </div>
      <span
         class="text-[11px] px-2.5 py-1 uppercase tracking-wider font-semibold rounded-md border {getStatusStyles(job.status)}"
      >
         {job.status}
      </span>
   </div>

   <!-- Card Body -->
   <div class="p-5 flex-1 flex flex-col gap-5 text-sm">
      <div class="flex justify-between items-center">
         <span class="text-gray-500">Created</span>
         <span class="text-gray-300">{formatDate(job.createdAt)}</span>
      </div>

      <div class="space-y-2">
         <span class="text-gray-500 text-xs uppercase tracking-wider"> Payload </span>
         <pre
            class="bg-black/50 p-3 text-xs text-gray-400 font-mono overflow-x-auto rounded-lg border border-white/5"
         >
{job.payload}
         </pre>
      </div>

      {#if job.result}
         <div class="space-y-2">
            <span class="text-gray-500 text-xs uppercase tracking-wider"> Result </span>
            <pre
               class="bg-black/50 p-3 text-xs text-purple-400 font-mono overflow-x-auto rounded-lg border border-purple-500/20"
            >
{job.result}
            </pre>
         </div>
      {/if}

      {#if job.error}
         <div class="space-y-2">
            <span class="text-rose-500/80 text-xs uppercase tracking-wider"> Error </span>
            <pre
               class="bg-rose-950/30 p-3 text-xs text-rose-400 font-mono overflow-x-auto rounded-lg border border-rose-500/20"
            >
{job.error}
            </pre>
         </div>
      {/if}
   </div>

   <!-- Card Actions -->
   <div class="p-5 flex gap-3 border-t border-white/5 mt-4 pt-4">
      {#if job.status === "idle"}
         <button
            onclick={() => onProcess(job.id)}
            disabled={isProcessing}
            class="flex-1 bg-white/10 text-white py-2.5 rounded-lg text-sm font-medium hover:bg-white/20 transition-colors disabled:opacity-50"
         >
            {isProcessing ? "Processing..." : "Process"}
         </button>
      {/if}
      <button
         onclick={() => onRemove(job.id)}
         class="flex-none px-4 py-2.5 text-gray-500 text-sm font-medium hover:text-white transition-colors"
      >
         Remove
      </button>
   </div>
</div>
