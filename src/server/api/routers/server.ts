import { z } from "zod";
import { Server, ServerSchema } from "~/lib/server";

import { createTRPCRouter, publicProcedure } from "~/server/api/trpc";

const servers: Server[] = [
  {
    id: 1,
    name: "Server 1",
    settings: {host: '127.0.0.1'},
    startType
: 
"wol"
  },
];

export const serverRouter = createTRPCRouter({
  update: publicProcedure.input(ServerSchema).mutation(async ({ input }) => {
    if (input.id <= 0) {
      input.id = servers.length + 1;
      servers.push(input);
      return input;
    }

    const index = servers.findIndex((s) => s.id === input.id);
    if (index === -1) throw new Error("Server not found");

    if (!servers[index]) throw new Error("Server not found");

    servers[index] = {
      ...servers[index],
      ...input,
      id: servers[index].id,
      lastChecked: new Date(),
    };

    return servers[index];
  }),

  delete: publicProcedure
    .input(z.object({ id: z.number() }))
    .mutation(async ({ input }) => {
      const index = servers.findIndex((s) => s.id === input.id);
      if (index === -1) throw new Error("Server not found");

      const deletedServer = servers.splice(index, 1)[0];

      return deletedServer;
    }),

  getAll: publicProcedure.query(() => {
    return servers;
  }),
});
