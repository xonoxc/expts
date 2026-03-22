import { sqliteTable, text } from "drizzle-orm/sqlite-core"

export const jobs = sqliteTable("jobs", {
   id: text("id").primaryKey(),
   status: text("status").default("idle"),
   payload: text("payload").notNull(),
   result: text("result"),
   error: text("error"),
   createdAt: text("created_at").notNull(),
   updatedAt: text("updated_at").notNull(),
})

export type Job = typeof jobs.$inferSelect
