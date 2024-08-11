import { z } from "zod";

export const ServerSchema = z.object({
  id: z.number().int(),
  name: z.string().min(1),
  host: z.string().min(1),
  type: z.string().min(1),
  isRunning: z.boolean().optional(),
  lastSeen: z.date().optional(),
  lastChecked: z.date().optional(),
});

export type Server = z.infer<typeof ServerSchema>;
