
<script lang="ts">
   import { createJobStore } from "./job.store"
   import JobCard from "./JobCard.svelte"

	const {
		handleProcessJob,
		handleProcessAll,
		disconnect,

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
	} = createJobStore()
</script>

<div class="min-h-screen bg-zinc-950 text-gray-200 p-8 md:p-12 relative overflow-hidden">
	<div
		class="absolute top-0 left-1/2 -translate-x-1/2 w-125 h-93.75 bg-purple-500/10 rounded-full blur-[120px] pointer-events-none -z-10 mix-blend-screen"
	></div>

	<div class="max-w-6xl mx-auto relative z-10">
		<div class="flex flex-col items-center text-center mb-20 pt-10">
			<div
				class="inline-flex items-center gap-3 px-4 py-1.5 rounded-full border border-white/10 bg-white/5 text-sm text-gray-300 mb-8 backdrop-blur-sm"
			>
				<span class="relative flex h-2 w-2">
					{#if connectionStatus === "connected"}
						<span
							class="animate-ping absolute inline-flex h-full w-full rounded-full {getConnectionDotColor(
								connectionStatus
							)} opacity-75"
						></span>
					{/if}
					<span
						class="relative inline-flex rounded-full h-2 w-2 {getConnectionDotColor(
							connectionStatus
						)}"
					></span>
				</span>
				<span class="capitalize tracking-wide">{connectionStatus} (Svelte)</span>
			</div>

			<h2 class="text-5xl md:text-6xl font-medium text-white tracking-tight mb-6">
				Job Monitor
			</h2>
			<p class="text-gray-400 text-lg max-w-xl mx-auto mb-10">
				Create, process, and track your background tasks in real-time.
			</p>

			<div class="flex flex-col sm:flex-row gap-4 justify-center w-full">
				<button
					onclick={handleCreateJob}
					disabled={isCreating}
					class="bg-purple-500 text-white px-8 py-3 rounded-lg font-medium hover:bg-purple-600 transition-all disabled:opacity-50 min-w-45 shadow-[0_0_20px_rgba(168,85,247,0.3)] hover:shadow-[0_0_30px_rgba(168,85,247,0.5)]"
				>
					{isCreating ? "Creating..." : "Start creating"}
				</button>

				<button
					onclick={handleCreateRandomJobs}
					disabled={isCreating}
					class="bg-transparent border border-white/10 text-white px-8 py-3 rounded-lg font-medium hover:bg-white/5 transition-colors disabled:opacity-50"
				>
					Explore random jobs
				</button>

				{#if idleJobsCount > 0}
					<button
						onclick={handleProcessAll}
						disabled={isProcessingAll}
						class="bg-emerald-600 text-white px-8 py-3 rounded-lg font-medium hover:bg-emerald-500 transition-all disabled:opacity-50 shadow-[0_0_20px_rgba(5,150,105,0.3)]"
					>
						{isProcessingAll ? "Processing..." : `Process all (${idleJobsCount})`}
					</button>
				{/if}
			</div>
		</div>

		{#if jobs.length > 0}
			<div class="space-y-8">
				<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
					{#each jobs as job (job.id)}
						<JobCard
							{job}
							onProcess={handleProcessJob}
							onRemove={handleRemoveJob}
							isProcessing={isProcessing === job.id}
						/>
					{/each}
				</div>

				<div class="flex justify-center pt-12">
					<button
						onclick={disconnect}
						class="text-gray-500 hover:text-white px-8 py-3 text-sm font-medium transition-colors border-b border-transparent hover:border-white/20"
					>
						Stop Monitoring
					</button>
				</div>
			</div>
		{:else}
			<div class="mt-12"></div>
		{/if}
	</div>
</div>
k
