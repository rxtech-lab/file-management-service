import { auth } from "@/auth";
import { NextRequest } from "next/server";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

// Force dynamic rendering to support streaming
export const dynamic = "force-dynamic";

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ fileId: string }> }
) {
  const { fileId } = await params;
  console.log(`[agent-stream] Starting connection for file ${fileId}`);

  const session = await auth();
  if (!session?.accessToken) {
    console.log("[agent-stream] No access token");
    return new Response(JSON.stringify({ error: "Unauthorized" }), {
      status: 401,
      headers: { "Content-Type": "application/json" },
    });
  }

  const url = `${API_BASE_URL}/api/files/${fileId}/agent-stream`;
  console.log(`[agent-stream] Fetching: ${url}`);

  try {
    const response = await fetch(url, {
      headers: {
        Authorization: `Bearer ${session.accessToken}`,
        Accept: "text/event-stream",
      },
      // Prevent Next.js from caching the response
      cache: "no-store",
    });

    console.log(`[agent-stream] Backend response: ${response.status}`);

    if (!response.ok) {
      const text = await response.text();
      console.log(`[agent-stream] Backend error: ${text}`);
      return new Response(
        JSON.stringify({ error: `Backend error: ${response.status}` }),
        {
          status: response.status,
          headers: { "Content-Type": "application/json" },
        }
      );
    }

    if (!response.body) {
      return new Response(JSON.stringify({ error: "No response body" }), {
        status: 502,
        headers: { "Content-Type": "application/json" },
      });
    }

    // Create a TransformStream to pipe data through without buffering
    const { readable, writable } = new TransformStream();

    // Pipe the backend response to the client
    response.body.pipeTo(writable).catch((err) => {
      console.error("[agent-stream] Pipe error:", err);
    });

    return new Response(readable, {
      headers: {
        "Content-Type": "text/event-stream",
        "Cache-Control": "no-cache, no-transform",
        Connection: "keep-alive",
        "X-Accel-Buffering": "no",
      },
    });
  } catch (error) {
    console.error("[agent-stream] Fetch error:", error);
    return new Response(
      JSON.stringify({
        error: error instanceof Error ? error.message : "Connection failed",
      }),
      { status: 502, headers: { "Content-Type": "application/json" } }
    );
  }
}
