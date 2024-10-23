import * as z from "zod"
import * as imports from "../null"
import { CompleteAction, RelatedActionModel } from "./index"

export const ServerModel = z.object({
  id: z.number().int(),
  name: z.string(),
  startTypeId: z.number().int().nullish(),
  stopTypeId: z.number().int().nullish(),
  checkTypeId: z.number().int().nullish(),
  isRunning: z.boolean().nullish(),
  lastSeen: z.date().nullish(),
  lastChecked: z.date().nullish(),
  settings: z.string(),
})

export interface CompleteServer extends z.infer<typeof ServerModel> {
  startType?: CompleteAction | null
  stopType?: CompleteAction | null
  checkType?: CompleteAction | null
}

/**
 * RelatedServerModel contains all relations on your model in addition to the scalars
 *
 * NOTE: Lazy required in case of potential circular dependencies within schema
 */
export const RelatedServerModel: z.ZodSchema<CompleteServer> = z.lazy(() => ServerModel.extend({
  startType: RelatedActionModel.nullish(),
  stopType: RelatedActionModel.nullish(),
  checkType: RelatedActionModel.nullish(),
}))
