"use server";

import { triggerOrganizeFile, getAgentStatus } from "@/lib/api/agent";
import type { OrganizeFileResponse, AgentStatusResponse } from "@/lib/api/types";

export async function organizeFileAction(
  fileId: number
): Promise<{ success: boolean; data?: OrganizeFileResponse; error?: string }> {
  try {
    const data = await triggerOrganizeFile(fileId);
    return { success: true, data };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to organize file",
    };
  }
}

export async function getAgentStatusAction(): Promise<{
  success: boolean;
  data?: AgentStatusResponse;
  error?: string;
}> {
  try {
    const data = await getAgentStatus();
    return { success: true, data };
  } catch (error) {
    return {
      success: false,
      error:
        error instanceof Error ? error.message : "Failed to get agent status",
    };
  }
}
