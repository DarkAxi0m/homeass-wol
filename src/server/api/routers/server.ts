import { z } from "zod";
import { Server, ServerSchema } from "~/lib/server";

import { createTRPCRouter, publicProcedure } from "~/server/api/trpc";

import { readFileSync, writeFileSync } from "fs";
import * as yaml from "js-yaml";

let servers: Server[] = [];
let loaded: boolean = false;

function loadYamlFile(filePath: string): any {
  try {
    const fileContents = readFileSync(filePath, "utf8");
    return yaml.load(fileContents);
  } catch (e) {
    console.error("Failed to load YAML file:", e);
  }
}

// Function to save YAML data to a file
function saveYamlFile(filePath: string, data: any): void {
  try {
    const yamlContent = yaml.dump(data);
    writeFileSync(filePath, yamlContent, "utf8");
  } catch (e) {
    console.error("Failed to save YAML file:", e);
  }
}

function load() {
  if (loaded) return;
  let s = loadYamlFile("servers.yaml");
  if (s) servers = s;
  loaded = true;
}

function save() {
  saveYamlFile("servers.yaml", servers);
}

export const serverRouter = createTRPCRouter({
  update: publicProcedure.input(ServerSchema).mutation(async ({ input }) => {
    if (input.id <= 0) {
      input.id = servers.length + 1;
      servers.push(input);
      save();
      return input;
    }

    const index = servers.findIndex((s) => s.id === input.id);
    if (index === -1) throw new Error("Server not found");

    if (!servers[index]) throw new Error("Server not found");

    servers[index] = {
      ...servers[index],
      ...input,
      id: servers[index].id,
 //     lastChecked: new Date(),
    };

    save();
    return servers[index];
  }),

  delete: publicProcedure
    .input(z.object({ id: z.number() }))
    .mutation(async ({ input }) => {
      const index = servers.findIndex((s) => s.id === input.id);
      if (index === -1) throw new Error("Server not found");

      const deletedServer = servers.splice(index, 1)[0];

      save();
      return deletedServer;
    }),

  getAll: publicProcedure.query(() => {
    load();
    return servers;
  }),
});
