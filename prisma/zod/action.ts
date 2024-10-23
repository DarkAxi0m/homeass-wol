import * as z from "zod"
import * as imports from "../null"
import { CompleteServer, RelatedServerModel } from "./index"

export const ActionModel = z.object({
  id: z.number().int(),
  name: z.string(),
  type: z.string(),
  reqValues: z.string(),
})

export interface CompleteAction extends z.infer<typeof ActionModel> {
  startServers: CompleteServer[]
  stopServers: CompleteServer[]
  checkServers: CompleteServer[]
}

/**
 * RelatedActionModel contains all relations on your model in addition to the scalars
 *
 * NOTE: Lazy required in case of potential circular dependencies within schema
 */
export const RelatedActionModel: z.ZodSchema<CompleteAction> = z.lazy(() => ActionModel.extend({
  startServers: RelatedServerModel.array(),
  stopServers: RelatedServerModel.array(),
  checkServers: RelatedServerModel.array(),
}))
