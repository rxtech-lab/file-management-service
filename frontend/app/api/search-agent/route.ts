import { streamText, convertToModelMessages, UIMessage, stepCountIs, createGateway, tool } from "ai";
import { createMCPClient } from "@ai-sdk/mcp";
import { z } from "zod";
import { auth } from "@/auth";
import { searchAgentPrompt } from "@/lib/prompts/search-agent";

export const maxDuration = 60;

// Schema for display_files tool
const displayFilesSchema = z.object({
  files: z.array(z.object({
    id: z.number().describe("File ID"),
    title: z.string().describe("File title"),
    description: z.string().optional().describe("Brief description of the file"),
    file_type: z.string().describe("Type of file: music, photo, video, document, invoice"),
    mime_type: z.string().optional().describe("MIME type"),
    folder_id: z.number().nullable().optional().describe("Parent folder ID"),
    folder_name: z.string().optional().describe("Parent folder name"),
  })),
  summary: z.string().optional().describe("Brief summary of search results to display before the list"),
});

type DisplayFilesInput = z.infer<typeof displayFilesSchema>;

// Client-side tool for displaying files in the UI
const displayFilesTool = {
  display_files: tool<DisplayFilesInput, DisplayFilesInput>({
    description: "Display ONLY the relevant files to the user based on their query. ALWAYS use this tool to show file search results. Do NOT include all search results - only include files that are most relevant to the user's specific request. Never include download links in your text responses.",
    inputSchema: displayFilesSchema,
    execute: async (input) => {
      return { files: input.files, summary: input.summary };
    },
  }),
};

// Create AI Gateway provider
const gateway = createGateway({
  apiKey: process.env.AI_GATEWAY_API_KEY,
});

export async function POST(req: Request) {
  const session = await auth();
  if (!session?.accessToken) {
    return Response.json({ error: "Unauthorized" }, { status: 401 });
  }

  let mcpClient: Awaited<ReturnType<typeof createMCPClient>> | null = null;

  try {
    const { messages }: { messages: UIMessage[] } = await req.json();

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

    const mcpTools = await mcpClient.tools();

    // Merge MCP tools with display_files tool
    const tools = { ...mcpTools, ...displayFilesTool };

    // Use model from env (format: "openai/gpt-5.2")
    const modelId = process.env.SEARCH_AGENT_MODEL || "openai/gpt-4o";

    const result = streamText({
      model: gateway(modelId),
      system: searchAgentPrompt,
      messages: await convertToModelMessages(messages),
      tools,
      stopWhen: stepCountIs(5), // Allow multiple tool calls and final text response
      onFinish: async () => {
        if (mcpClient) {
          await mcpClient.close();
        }
      },
    });

    return result.toUIMessageStreamResponse();
  } catch (error) {
    console.error("[search-agent] Error:", error);
    if (mcpClient) {
      await mcpClient.close();
    }
    return Response.json(
      { error: error instanceof Error ? error.message : "Internal server error" },
      { status: 500 }
    );
  }
}
