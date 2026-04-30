"use client"

export function UserMessage({ content }: { content: string }) {
  return (
    <div className="flex justify-end mb-4">
      <div className="bg-primary text-primary-foreground rounded-lg p-3 max-w-[80%]">
        {content}
      </div>
    </div>
  )
}
