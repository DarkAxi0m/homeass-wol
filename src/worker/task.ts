import { Server } from "~/lib/server";
import * as ping from "ping";

import { CronJob } from "cron";
async function updateServer(id: number, state: boolean, msg: string) {
  console.log(">>>", id, state);
  return (
    await fetch("http://127.0.0.1:3000/api/update/" + id, {
      method: "POST",
      body: JSON.stringify({
        id,
        state,
        msg,
      }),
      headers: {
        "Content-type": "application/json; charset=UTF-8",
      },
    })
  ).json();
}

async function getServers() {
    return await (await fetch("http://127.0.0.1:3000/api/trpc/server.getAll")).json();

}

async function main() {
  const servers = (await getServers()).result.data.json as Server[]

  servers.forEach(async (s: Server) => {
      s.settings = JSON.parse((<any> s.settings)as string)
    ping.promise.probe(s.settings.host).then(async function (res) {
      console.log(s.id, res.host, res.alive);
      console.log(await updateServer(s.id, res.alive, res.output));
    });
  });
}

console.log("started....");
const job = new CronJob(
  "0,30  * * * * *",
  function () {
    console.log("Checking");
    main();
  },
  null,
  true,
);

main();
