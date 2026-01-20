"use server";

import { createStreamableValue } from "@ai-sdk/rsc";
import { streamText } from "ai";
import { createMCPClient } from "@ai-sdk/mcp";
import { auth } from "@/auth";
import { searchAgentPrompt } from "@/lib/prompts/search-agent";
import type { AgentSearchProgress, AgentSearchMessage } from "@/lib/api/types";

function formatToolName(toolName: string): string {
  return toolName
    .split("_")
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
}

export async function searchWithAgentAction(messages: AgentSearchMessage[]) {
  const session = await auth();

  if (!session?.accessToken) {
    const errorStream = createStreamableValue<AgentSearchProgress>({
      status: "error",
      message: "Authentication required",
    });
    errorStream.done();
    return { progress: errorStream.value, textStream: null };
  }

  const progressStream = createStreamableValue<AgentSearchProgress>({
    status: "idle",
    message: "Initializing...",
  });

  const textStream = createStreamableValue<string>("");

  // Run async to allow returning stream immediately
  (async () => {
    let mcpClient: Awaited<ReturnType<typeof createMCPClient>> | null = null;

    try {
      progressStream.update({
        status: "thinking",
        message: "Connecting to search service...",
      });

      const url = process.env.NEXT_PUBLIC_API_URL + "/mcp";

      mcpClient = await createMCPClient({
        transport: {
          type: "http",
          url: url,
          headers: {
            Authorization: `Bearer ${session.accessToken}`,
          },
        },
      });

      progressStream.update({
        status: "thinking",
        message: "Preparing search...",
      });

      const tools = await mcpClient.tools();

      // Convert messages to AI SDK format
      const aiMessages = messages.map((msg) => ({
        role: msg.role as "user" | "assistant",
        content: msg.content,
      }));

      const result = streamText({
        model: process.env
          .SEARCH_AGENT_MODEL! as Parameters<typeof streamText>[0]["model"],
        system: searchAgentPrompt,
        messages: aiMessages,
        tools: tools,
        onChunk: ({ chunk }) => {
          if (chunk.type === "tool-call") {
            progressStream.update({
              status: "calling",
              toolName: chunk.toolName,
              toolArgs: (chunk as { input?: unknown }).input as Record<string, unknown>,
              message: `${formatToolName(chunk.toolName)}...`,
            });
          }
          if (chunk.type === "tool-result") {
            progressStream.update({
              status: "calling",
              toolName: (chunk as { toolName?: string }).toolName,
              toolResult: (chunk as { result?: unknown }).result,
              message: `Processing results...`,
            });
          }
        },
      });

      // Stream the text response
      for await (const textPart of result.textStream) {
        textStream.append(textPart);
      }

      progressStream.done({
        status: "complete",
        message: "Search complete",
      });
      textStream.done();
    } catch (error) {
      console.error("Search agent error:", error);
      progressStream.done({
        status: "error",
        message:
          error instanceof Error ? error.message : "Failed to search files",
      });
      textStream.done();
    } finally {
      if (mcpClient) {
        await mcpClient.close();
      }
    }
  })();

  return { progress: progressStream.value, textStream: textStream.value };
}
