import { PrismaClient } from '@prisma/client';

const prisma = new PrismaClient();

async function main() {
  await prisma.action.createMany({
    data: [
      {
        name: "wol",
        type: "start",
        reqValues: JSON.stringify({ host: "string", MAC: "string" }),
      },
      {
        name: "ssh",
        type: "stop",
        reqValues: JSON.stringify({ host: "string", username: "string" }),
      },
      {
        name: "ping",
        type: "check",
        reqValues: JSON.stringify({ host: "string" }),
      },
      {
        name: "connect",
        type: "check",
        reqValues: JSON.stringify({ host: "string", port: "number" }),
      },
    ],
  });
}

main()
  .catch((e) => {
    console.error(e);
    process.exit(1);
  })
  .finally(async () => {
    await prisma.$disconnect();
  });

