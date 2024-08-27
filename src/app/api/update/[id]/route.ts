import { type NextRequest } from "next/server";
import { NextResponse } from "next/server";
import { api } from "~/trpc/server";

const handler = async (
  req: NextRequest,
  { params }: { params: { slug: string } },
) => {
  const res = await req.json();

  console.log("-------------------------------------");
  console.log(res);
  console.log("-------------------------------------");

  const ures = await api.server.checked(res);
  return NextResponse.json({ params, ures });
};

export { handler as GET, handler as POST };
