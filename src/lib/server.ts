import { z } from "zod";

export const ServerSchema = z.object({
  id: z.number().int(),
  name: z.string().min(1),
  startType: z.string().optional(),
  stopType: z.string().optional(),
  checkType: z.string(),
  isRunning: z.boolean().optional(),
  lastSeen: z.date().optional(),
  lastChecked: z.date().optional(),
  settings: z.record(z.string()),
});

export type Server = z.infer<typeof ServerSchema>;

const ActionTypeSchema = z.enum(["stop", "start", "check"]);

export const ActionSchema = z.object({
  name: z.string().min(1),
  type: ActionTypeSchema,
  reqValues: z.map(z.string(), z.string()),
});

export type Action = z.infer<typeof ActionSchema>;

const HostName = { host: "string" };

export const Actions: Record<string, Action> = {
  wol: {
    name: "wol",
    type: "start",
    reqValues: { ...HostName, ...{ MAC: "string" } },
  },
  ssh: {
    name: "ssh",
    type: "stop",
    reqValues: { ...HostName, ...{ username: "string" } },
  },
  ping: { name: "ping", type: "check", reqValues: HostName },
  connect: {
    name: "connect",
    type: "check",
    reqValues: { ...HostName, ...{ port: "number" } },
  },
};

export function ActionsByType(type: string): Action[] {
  return Object.values(Actions).filter((action) => action.type === type);
}
