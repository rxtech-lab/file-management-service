import { apiClient } from "./client";
import type { AgentStatusResponse, OrganizeFileResponse } from "./types";

export async function triggerOrganizeFile(
  fileId: number
): Promise<OrganizeFileResponse> {
  return apiClient<OrganizeFileResponse>(`/api/files/${fileId}/organize`, {
    method: "POST",
  });
}

export async function getAgentStatus(): Promise<AgentStatusResponse> {
  return apiClient<AgentStatusResponse>("/api/agent/status");
}

export function createAgentEventSource(fileId: number): EventSource {
  // Use Next.js API route to proxy the SSE stream with authentication
  return new EventSource(`/api/agent-stream/${fileId}`);
}
