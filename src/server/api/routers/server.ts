import { ServerModel } from "prisma/zod";
import { z } from "zod";

import { createTRPCRouter, publicProcedure } from "~/server/api/trpc";

import { shutdownServer } from "~/worker/SSH";



export const serverRouter = createTRPCRouter({
 
  update: publicProcedure.input(ServerModel)
  .mutation(async ({ input, ctx }) => {
    const { id, ...dataWithoutId } = input;
    if (id <= 0) {
        console.log('Saving new server', input)
       return await ctx.db.server.create({
       data: dataWithoutId}) 
    }

     console.log('Saving server', id, dataWithoutId)

    return await ctx.db.server.update({
               data: dataWithoutId,
               where: {id}

    })
  }),

  delete: publicProcedure
    .input(z.object({ id: z.number() }))
  .mutation(async ({ input, ctx }) => {
  return await ctx.db.server.delete({where:{id:input.id}})  
    }),

  getAll: publicProcedure.query(({ctx}) => {
    return ctx.db.server.findMany();
  }),

 actions: publicProcedure.query(({ctx}) => {
     return ctx.db.action.findMany();
 }),

 checked: publicProcedure.input(z.object(
     {
        id: z.number(),
        state: z.boolean(),
        msg: z.string()
     }
 )
 ).mutation(async ({input, ctx})=>{
console.log(

    
    '***************',input.id)

    return await ctx.db.server.update(
    {
        data: {
             lastSeen: (input.state? new Date():null),
             lastChecked: new Date()
        },
        where: {id: input.id}
    }

    )

    }),

  stop: publicProcedure
    .input(
      z.object({
        id: z.number(),
      }),
    )
    .mutation(async ({ input }) => {
   /*   const index = servers.findIndex((s) => s.id === input.id);
      if (index === -1)
        throw new Error("Server not found, " + index + ", " + input.id);

      if (!servers[index]) throw new Error("Server not found");
      const s = servers[index];
      console.log(servers[index]);

      shutdownServer({
        ...s.settings,
        ...{ privateKeyPath: "/home/chris/.ssh/id_rsa" },
      });

      return input;*/
    }),
});
