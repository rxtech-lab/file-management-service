import { auth } from "@/auth";
import { NextRequest } from "next/server";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ fileId: string }> }
) {
  const session = await auth();
  const { fileId } = await params;

  if (!session?.accessToken) {
    return new Response(JSON.stringify({ error: "Unauthorized" }), {
      status: 401,
      headers: { "Content-Type": "application/json" },
    });
  }

  // Proxy the SSE request to the backend with auth
  const response = await fetch(
    `${API_BASE_URL}/api/files/${fileId}/agent-stream`,
    {
      headers: {
        Authorization: `Bearer ${session.accessToken}`,
        Accept: "text/event-stream",
      },
    }
  );

  if (!response.ok) {
    return new Response(
      JSON.stringify({ error: `Backend error: ${response.status}` }),
      {
        status: response.status,
        headers: { "Content-Type": "application/json" },
      }
    );
  }

  // Stream the response back to the client
  return new Response(response.body, {
    headers: {
      "Content-Type": "text/event-stream",
      "Cache-Control": "no-cache",
      Connection: "keep-alive",
    },
  });
}
